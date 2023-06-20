import sys

import pyqtgraph.parametertree
from PySide6 import QtWidgets, QtCore
from PySide6.QtCore import QDir, Qt
from PySide6.QtWidgets import (
    QApplication,
    QWidget,
    QMainWindow,
    QTreeView,
    QDockWidget,
)


class MainApp(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("Hey there")

        ptree = PacketTree(self)
        self.setCentralWidget(ptree)


class PacketTree(QWidget):
    """PacketView is a widget that shows a tree of packets as well as properties on them when selected."""

    def __init__(self, parent: QtWidgets.QWidget | None = None):
        super().__init__(parent)
        self.setWindowTitle("Packet Overview")
        splitter = QtWidgets.QSplitter(self)
        layout = QtWidgets.QVBoxLayout()

        splitter.setOrientation(Qt.Vertical)
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
