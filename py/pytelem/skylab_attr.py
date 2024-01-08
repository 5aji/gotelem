
from typing import Dict, Optional, List, Union
from enum import Enum
from attrs import define, field, validators


@define
class Bus():
    name: str
    baud_rate: str
    extended_id: bool = False




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


@define
class CustomTypeDef():
    name: str
    base_type: FieldType # should be a strict size
    values: Union[List[str], Dict[str, int]]


@define
class BitfieldBit():
    "micro class to represent one bit in bitfields"
    name: str

@define
class Field():
    name: str = field(validator=[validators.matches_re(r"^[A-Za-z0-9_]+$")])
    type: FieldType

    #metadata
    units: Optional[str]
    conversion: Optional[float]


@define
class BitField():
    name: str = field(validator=[validators.matches_re(r"^[A-Za-z0-9_]+$")])
    type: str = field(default="bitfield", init=False) # it's a constant value
    bits: List[BitfieldBit]


class Endian(str, Enum):
    BIG = "big"
    LITTLE = "little"

@define
class Packet():
    name: str
    description: str
    id: int
    endian: Endian
    frequency: Optional[int]
    data: List[Field]

@define
class RepeatedPacket():
    name: str
    description: str
    id: int
    endian: Endian
    frequency: Optional[int]
    data: List[Field]
    repeat: int
    offset: int


