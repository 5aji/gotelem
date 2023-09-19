import sys
import logging

from PySide6.QtCore import QObject, Slot, Signal
from PySide6.QtWidgets import QPlainTextEdit


class Bridge(QObject):
    log = Signal(str)


class QLogHandler(logging.Handler):
    bridge = Bridge()

    def __init__(self):
        super().__init__()

    def emit(self, record):
        msg = self.format(record)
        self.bridge.log.emit(msg)
