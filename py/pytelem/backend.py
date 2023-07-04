import aiohttp
import orjson
import threading

# connect to websocket - create thread that handles JSON events
class TelemetryServer(QObject):
    """Connection to upstream database"""

    conn_url: str
    "Something like http://<some_ip>:8082"

    callbacks: Dict[str, Signal]

    def __init__(self, url: str, parent=None):
        super().__init__(parent)
        self.conn_url = url

        

