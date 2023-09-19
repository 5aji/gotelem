import random
import sys
import logging

import pyqtgraph.parametertree
from PySide6 import QtWidgets, QtCore
from PySide6.QtCore import QDir, Qt, QObject, Slot, Signal, QTimer
from PySide6.QtGui import QAction
from PySide6.QtWidgets import (
    QApplication,
    QWidget,
    QMainWindow,
    QTreeView,
    QDockWidget, QToolBar, QPlainTextEdit,
)
from gui_log import QLogHandler
from pytelem.widgets.smart_display import SmartDisplay
from bms import BMSOverview


class DataStore(QObject):
    """Stores all packets and timestamps for display and logging.
    Queries the upstreams for the packets as they come in as well as historical"""

    def __init__(self, remote):
        super().__init__()


class MainApp(QMainWindow):
    new_data = Signal(float)

    def __init__(self):
        super().__init__()
        self.setWindowTitle("pyview")
        layout = QtWidgets.QVBoxLayout()

        mb = self.menuBar()
        self.WindowMenu = mb.addMenu("Windows")

        bms = BMSOverview()
        packet_tree = QDockWidget('Packet Tree', self)
        self.addDockWidget(Qt.DockWidgetArea.LeftDockWidgetArea, packet_tree)
        packet_tree.setWidget(PacketTreeView())
        packet_tree.hide()
        self.ShowPacketTree = packet_tree.toggleViewAction()
        self.WindowMenu.addAction(self.ShowPacketTree)

        log_dock = QDockWidget('Application Log', self)
        self.qlogger = QLogHandler()
        self.log_box = QPlainTextEdit()
        self.log_box.setReadOnly(True)
        log_dock.setWidget(self.log_box)
        self.qlogger.bridge.log.connect(self.log_box.appendPlainText)
        self.addDockWidget(Qt.DockWidgetArea.BottomDockWidgetArea, log_dock)

        self.logger = logging.Logger("Main")
        self.logger.addHandler(self.qlogger)
        self.logger.info("hi there!")
        self.ShowLog = log_dock.toggleViewAction()
        self.ShowLog.setShortcut("CTRL+L")
        self.WindowMenu.addAction(self.ShowLog)
        self.display = SmartDisplay(self, "test")
        self.new_data.connect(self.display.update_value)
        # start a qtimer to generate random data.
        self.timer = QTimer(parent=self)
        self.timer.timeout.connect(self.__random_data)
        # self.__random_data.connect(self.timer.timeout)
        self.timer.start(100)

        self.setCentralWidget(self.display)

    @Slot()
    def __random_data(self):
        # emit random data to the new_data
        yay = random.normalvariate(10, 1)
        self.logger.info(yay)
        self.new_data.emit(yay)


class PacketTreeView(QWidget):
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


class SolverView(QWidget):
    """Main Solver Widget/Window"""


if __name__ == "__main__":
    app = QApplication(sys.argv)
    main_window = MainApp()
    main_window.show()
    app.exec()
