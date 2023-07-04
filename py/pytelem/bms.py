from typing_extensions import TypedDict

from PySide6.QtCore import QObject
from PySide6.QtWidgets import QDockWidget, QGridLayout, QGroupBox, QHBoxLayout, QLabel, QVBoxLayout, QWidget

ContactorStates = TypedDict("ContactorStates", {})


class BMSState(QObject):
    """Represents the BMS state, including history."""
    main_voltage: float
    aux_voltage: float
    current: float

    def __init__(self, parent=None, upstream=None):
        super().__init__(parent)
        # uhh, take a connection to the upstream?


class BMSModuleViewer(QDockWidget):
    """BMS module status viewer (temp and voltage)"""

    layout: QGridLayout

    def __init__(self) -> None:
        super().__init__()
        self.layout = QGridLayout(self)


class BMSOverview(QWidget):
    layout: QGridLayout


    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.layout = QGridLayout(self)
        main_voltage_label = QLabel("Main Voltage", self)
        self.layout.addWidget(main_voltage_label, row=1, column=0)
        aux_v_label= QLabel("Aux Voltage", self)
        self.layout.addWidget(aux_v_label, row=1, column=1)
        current_label = QLabel("Battery Current", self)
        self.layout.addWidget(current_label, row=1, column=2)

        # now add widgets that display the numeric values.
        # then make slots that take floats and display them.

class BMSStatus(QDockWidget):

    layout: QVBoxLayout
    contactors: QGroupBox

    def __init__(self, parent = None):
        super().__init__("Battery Status", parent)

        self.layout = QVBoxLayout(self)
        self.contactors = QGroupBox("Contactor State", self)
        self.layout.addWidget(self.contactors)

