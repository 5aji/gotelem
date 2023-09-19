from functools import cached_property

import aiohttp
import orjson
import threading
from typing import Dict

from PySide6.QtCore import QObject, Signal, Slot

from pytelem.skylab import SkylabFile


# connect to websocket - create thread that handles JSON events
class TelemetryServer(QObject):
    """Connection to upstream database"""

    conn_url: str
    "Something like http://<some_ip>:8082"

    def __init__(self, url: str, parent=None):
        super().__init__(parent)
        self.conn_url = url

    NewPacket = Signal(object)
    """Signal that is emitted when a new packet is received in realtime. Contains the packet itself"""

    @cached_property
    def schema(self) -> SkylabFile:
        """Gets the Packet Schema from the server"""
        pass

    @Slot()
    def connect(self):
        """Attempt to connect to server"""

    def query(self, queryparams):
        """Query the historical data and store the result in the datastore"""


