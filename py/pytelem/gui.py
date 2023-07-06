import sys
import logging

import pyqtgraph.parametertree
from PySide6 import QtWidgets, QtCore
from PySide6.QtCore import QDir, Qt, QObject
from PySide6.QtWidgets import (
    QApplication,
    QWidget,
    QMainWindow,
    QTreeView,
    QDockWidget,
)

from bms import BMSOverview

class QtLogger(logging.Handler, QObject):
    appendLog = QtCore.Signal(str)
    
    def __init__(self, parent):
        super().__init__()
        QtCore.QObject.__init__(self)
        self.widget = QtWidgets.QPlainTextEdit(parent)
        self.widget.setReadOnly(True)
        self.appendLog.connect(self.widget.appendPlainText)
    
    def emit(self, record):
        msg = self.format(record)
        self.appendLog.emit(msg)


class DataStore:
    """Stores all packets and timestamps for display and logging.
    Queries the upstreams for the packets as they come in as well as historical"""
    
    def __init__(self, remote):
        pass

class MainApp(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("Hey there")
        layout = QtWidgets.QVBoxLayout()

        bms = BMSOverview()
        dw = QDockWidget('bms', self)
        self.addDockWidget(Qt.DockWidgetArea.LeftDockWidgetArea, dw)
        dw.setWidget(PacketTree())
        self.setCentralWidget(bms)



class PacketTree(QWidget):
    """PacketView is a widget that shows a tree of packets as well as properties on them when selected."""

    def __init__(self, parent: QtWidgets.QWidget | None = None):
        super().__init__(parent)
        self.setWindowTitle("Packet Overview")
        splitter = QtWidgets.QSplitter(self)
        layout = QtWidgets.QVBoxLayout()

#        splitter.setOrientation(Qt.Vertical)
        self.tree = QTreeView()
        self.prop_table = pyqtgraph.parametertree.ParameterTree()
        splitter.addWidget(self.tree)
        splitter.addWidget(self.prop_table)
        layout.addWidget(splitter)

        self.setLayout(layout)


if __name__ == "__main__":
    app = QApplication(sys.argv)
    main_window = MainApp()
    main_window.show()
    app.exec()
