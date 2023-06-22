# this file describes a skylab yaml and it's associated fields. It also
# provides functions to parse a skylab packet folder and a few AST operators.
from abc import ABC, abstractmethod
import re
from pathlib import Path
from typing import Callable, Iterable, NewType, TypedDict, List, Protocol, Union, Set

from pydantic import BaseModel, validator
from enum import Enum
import yaml
import jinja2


# This part of the file is dedicated to parsing the skylab yaml files. We define
# classes that represent objects in the yaml files, and perform basic validation on
# the input data. We also define a load_yamls function that loads a directory of skylab files.


class FieldType(str, Enum):
    """FieldType indicates the type of the field - the enum represents the C type,
    but you can use a map to convert the type to another language."""

    # used to ensure types are valid, and act as representations for other languages/mappings.
    U8 = "uint8_t"
    U16 = "uint16_t"
    U32 = "uint32_t"
    U64 = "uint64_t"
    I8 = "int8_t"
    I16 = "int16_t"
    I32 = "int32_t"
    I64 = "int64_t"
    F32 = "float"

    Bitfield = "bitfield"

    def size(self) -> int:
        """Returns the size, in bytes, of the type."""
        match self:
            case FieldType.U8:
                return 1
            case FieldType.U16:
                return 2
            case FieldType.U32:
                return 4
            case FieldType.U64:
                return 8
            case FieldType.I8:
                return 1
            case FieldType.I16:
                return 2
            case FieldType.I32:
                return 4
            case FieldType.I64:
                return 8
            case FieldType.F32:
                return 4
            case FieldType.Bitfield:
                return 1
        return -1


# A FieldTypeMapper is any function that takes a field type and returns
# either a mapped string (based on the type) or none if there was not a match.
FieldTypeMapper = Callable[[FieldType], str | None]


class _Bits(TypedDict):
    """Internal class: a bits object just has a name."""

    name: str


class SkylabField(BaseModel):
    """Represents a field (data element) inside a Skylab Packet."""

    name: str
    "the name of the field. must be alphanumeric and underscores"
    type: FieldType
    "the type of the field"
    units: str | None
    "optional descriptor of the unit representation"
    conversion: float | None
    "optional conversion factor to be applied when parsing"
    bits: List[_Bits] | None
    "if the type if a bitfield, "

    @validator("bits")
    def bits_must_exist_if_bitfield(cls, v, values):
        if v is None and "type" in values and values["type"] is FieldType.Bitfield:
            raise ValueError("bits are not present on bitfield type")
        if (
            v is not None
            and "type" in values
            and values["type"] is not FieldType.Bitfield
        ):
            raise ValueError("bits are present on non-bitfield type")
        return v

    @validator("name")
    def name_valid_string(cls, v: str):
        if not re.match(r"^[A-Za-z0-9_]+$", v):
            return ValueError("invalid name")
        return v

    @validator("name")
    def name_nonzero_length(cls, v: str):
        if len(v) == 0:
            return ValueError("name cannot be empty string")
        return v


class Endian(str, Enum):
    """Symbol representing the endianness of the packet"""

    Big = "big"
    Little = "little"


class SkylabPacket(BaseModel):
    """Represents a CAN packet. Contains SkylabFields with information on the structure of the data."""

    name: str
    description: str | None
    id: int
    endian: Endian
    repeat: int | None
    offset: int | None
    data: List[SkylabField]

    # @validator("data")
    # def packet_size_limit(cls, v: List[SkylabField]):
    #     tot = sum([f.type.size() for f in v])
    #     if tot > 8:
    #         return ValueError("Total packet size cannot exceed 8 bytes")
    #     return v

    @validator("id")
    def id_non_negative(cls, v: int) -> int:
        if v < 0:
            raise ValueError("id must be above zero")
        return v

    @validator("name")
    def name_valid_string(cls, v: str) -> str:
        if not re.match(r"^[A-Za-z0-9_]+$", v):
            raise ValueError("invalid name", v)
        return v

    @validator("name")
    def name_nonzero_length(cls, v: str) -> str:
        if len(v) == 0:
            raise ValueError("name cannot be empty string")
        return v

    @validator("offset")
    def offset_must_have_repeat(cls, v: int | None, values) -> int | None:
        if v is not None and "repeat" in values and values["repeat"] is not None:
            raise ValueError("field with offset must have repeat defined")
        return v

    @validator("repeat")
    def repeat_gt_one(cls, v: int | None):
        if v is not None and v <= 1:
            raise ValueError("repeat must be strictly greater than one")
        return v


class SkylabBoard(BaseModel):
    """Represents a single board. Each board has packets that it sends and receives

    Validations:
    - There can only be one sender of a packet, but multiple receivers
    - every name in the transmit/receive list must have a corresponding packet.
    """

    name: str
    "The name of the board"
    transmit: List[str]
    "The packets sent by this board"
    receive: List[str]
    "The packets received by this board."

    @validator("name")
    def name_valid_string(cls, v: str):
        if not re.match(r"^[A-Za-z0-9_]+$", v):
            return ValueError("invalid name", v)
        return v

    @validator("name")
    def name_nonzero_length(cls, v: str):
        if len(v) == 0:
            return ValueError("name cannot be empty string")
        return v


class SkylabBus(BaseModel):
    name: str
    "The name of the bus"
    baud_rate: int
    "Baud rate setting for the bus"
    extended_id: bool
    "If the bus uses extended ids"

    @validator("name")
    def name_valid_string(cls, v: str):
        if not re.match(r"^[A-Za-z0-9_]+$", v):
            return ValueError("invalid name", v)
        return v

    @validator("baud_rate")
    def baud_rate_supported(cls, v: int):
        if v not in [125000, 250000, 500000, 750000, 1000000]:
            raise ValueError("unsupported baud rate", v)
        return v


class SkylabFile(BaseModel):
    """Represents an entire skylab yaml file. Performs additional cross-validation between
    boards and packets."""

    packets: List[SkylabPacket] = []
    boards: List[SkylabBoard] = []
    busses: List[SkylabBus] = []

    # TODO: add extra validators here?


def load_skylab_dir(path: Path) -> SkylabFile:
    """Loads all the .yaml files in a directory and merges them into one large SkylabFile, which is then returned."""
    files = [f for f in path.iterdir() if re.search(r".*\.ya?ml$", str(f))]
    sky_files: List[SkylabFile] = []
    for file in files:
        with open(file, "r") as f:
            obj = yaml.load(f, Loader=yaml.Loader)
            sky_files.append(SkylabFile.parse_obj(obj))
    # merge the files

    # this is not very fast or elegant but who cares.
    all_pkts = []
    all_boards = []
    for sky_f in sky_files:
        d = sky_f.dict()
        all_pkts.append(d["packets"])
        all_boards.append(d["boards"])
    collected_skyfile = SkylabFile.parse_obj(
        {"packets": all_pkts, "boards": all_boards}
    )

    return collected_skyfile


# hey. The next bit of code is entirely optional for you to use! It's totally acceptable to just skip it and manually
# iterate over the SkylabFile yourself. The reason we use the walk tree/visitor pattern here is to abstract away
# traversing the tree from the functions that process it.

# While this is generally a pretty common use case (hence the abstraction), it
# can be difficult to wrap certain processes around it.


SkylabObject = Union[SkylabFile, SkylabPacket, SkylabBoard, SkylabField, SkylabBus]
"SkylabObject is any object that will be walked when making a parsing pass on the AST"


class SkylabWalker(Protocol):
    """A SkylabWalker is any class that implements walk(self, SkylabObject)."""

    def walk(self, obj: SkylabObject):
        ...

    "walk is called for each SkylabObject in the SkylabFile tree"


class SkylabVisitor(ABC):
    """SkylabVisitor is an abstract class that makes children walkable. Children must
    implement the visit_* functions which contain explicit signatures for each discrete unit in a Skylabfile.
    """

    @abstractmethod
    def visit_file(self, file: SkylabFile):
        ...

    @abstractmethod
    def visit_board(self, board: SkylabBoard):
        ...

    @abstractmethod
    def visit_packet(self, packet: SkylabPacket):
        ...

    @abstractmethod
    def visit_field(self, field: SkylabField, parent: SkylabPacket):
        ...

    @abstractmethod
    def visit_bus(self, bus: SkylabBus):
        ...

    _last_parent: SkylabPacket | None = None
    "internal variable storing the parent packet for visit_field"

    def walk(self, obj: SkylabObject):
        match obj:
            case SkylabFile():
                self.visit_file(obj)
            case SkylabBoard():
                self.visit_board(obj)
            case SkylabPacket():
                self._last_parent = obj
                self.visit_packet(obj)
            case SkylabBus():
                self.visit_bus(obj)
            case SkylabField():
                if self._last_parent is None:
                    raise Error("Unexpected field without parent")
                self.visit_field(obj, self._last_parent)


def walk_tree(tree: SkylabFile, walker: SkylabWalker):
    """Walks the tree using the given walker"""

    walker.walk(tree)

    for bus in tree.busses:
        walker.walk(bus)

    for board in tree.boards:
        walker.walk(board)

    for packet in tree.packets:
        walker.walk(packet)
        for field in packet.data:
            walker.walk(field)


class Error(Exception):
    """An exception when processing the tree"""


class RelationValidator:
    """This class is a processor that validates the relation between boards and
    packets.

    - Each packet MAY have AT MOST one transmitter.
    - Each packet MUST have AT LEAST one board reference.
    - Boards MUST ONLY reference packets that exist in 'packets'
    - Boards MUST have a UNIQUE name
    - Packets MUST have a UNIQUE name"""

    seen_packets: Set[str]
    "A set of all the packets in the 'packets' field"

    sent_packets: Set[str]
    "A set of all the packet names that are sent by boards"
    recv_packets: Set[str]
    "A set of all packets recv'd by boards"

    board_names: Set[str]

    def __init__(self):
        self.board_names = set()
        self.seen_packets = set()
        self.sent_packets = set()
        self.board_names = set()

    # The first test: make sure that no two boards send the same packet.
    # check sent_packets for existing element before adding.

    # the second test: each packet should have a unique name.

    # the third test -> Union of sent_packets and recv_packets should
    # be exactly equal to seen_packets

    def walk(self, obj: SkylabObject):
        match obj:
            case SkylabPacket(name=n):
                if n in self.seen_packets:
                    raise Error(f"packet {n} declared twice")
                self.seen_packets.add(n)
            case SkylabBoard(transmit=tx, receive=rx, name=n):
                if n in self.board_names:
                    raise Error(f"board {n} declared twice")
                self.board_names.add(n)
                for r_pkt in rx:
                    self.recv_packets.add(r_pkt)

                for t_pkt in tx:
                    if t_pkt in self.sent_packets:
                        raise Error(f"packet {t_pkt} is sent from two sources")
                    self.sent_packets.add(t_pkt)
            case _:
                ...  # skip others.

    def validate(self):
        """runs final checks"""
        # perform third check: packets must all be used.
        board_ref_packets = self.recv_packets.union(self.sent_packets)

        # xor = symmetric_difference, which is packets in one or the other but not both.
        unref_packets = board_ref_packets ^ self.seen_packets
        if len(unref_packets) > 0:
            raise Error(f"packets missing a link: {unref_packets}")


class CollisionDetector:
    """This class detects ID collisions of packets. It expands the repeated packet to ensure that there is
    no overlap"""

    seen_ids: Set[int] = set()

    def add_or_fail(self, idx: int):
        if idx in self.seen_ids:
            raise ValueError(f"Collision on packet {idx}")
        self.seen_ids.add(idx)

    def walk(self, obj: SkylabObject):
        match obj:
            case SkylabPacket(id=idx, offset=None, repeat=None):
                # matches single packets - just add it directly.
                self.add_or_fail(idx)

            # we need the guard clause cause pyright/mypy isn't smart enough to read the entire match block.
            case SkylabPacket(
                id=base_idx, offset=off, repeat=rpt
            ) if off is not None and rpt is not None:
                for i in range(0, rpt):
                    self.add_or_fail(base_idx + i * off)
            case _:
                ...  # do nothing for packets or fields.

    def validate(self):
        print(f"{len(self.seen_ids)} packet IDs discovered with no collisions")


class ExampleGenerator:
    """Demonstrates how to use Jinja templates with a custom environment to generate output documents from
    the skylab objects."""

    env: jinja2.Environment

    def __init__(self):
        self.env = jinja2.Environment(
            loader=jinja2.loaders.PackageLoader(".templates.c")
        )

    def render(self, skylab: SkylabFile):
        ...


class GraphvizGenerator:
    """This class converts the Skylab files into a GraphViz document
    detailing the flow of data as well as information about the data."""


class CGenerator:
    """This class generates C files for our microcontrollers."""


class PyGenerator:
    """This class generates a python module that can serialize/deserialize packets."""
