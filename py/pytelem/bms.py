from typing import Dict
from typing_extensions import TypedDict

from PySide6.QtCore import QObject, Qt, Slot
from PySide6.QtGui import QFontDatabase
from PySide6.QtWidgets import QButtonGroup, QDockWidget, QGridLayout, QGroupBox, QHBoxLayout, QLabel, QRadioButton, QVBoxLayout, QWidget

ContactorStates = TypedDict("ContactorStates", {})


class BMSState(QObject):
    """Represents the BMS state, including history."""
    main_voltage: float
    aux_voltage: float
    current: float

    def __init__(self, parent=None, upstream=None):
        super().__init__(parent)
        # uhh, take a connection to the upstream?


class BMSModuleViewer(QWidget):
    """BMS module status viewer (temp and voltage)"""

    # use graphics view for rendering.

    temps: list[float] = []
    volts: list[float] = []

    def __init__(self, parent = None) -> None:
        super().__init__(parent)

        layout = QGridLayout()

        bg = QButtonGroup(self)

        self.volts_btn = QRadioButton("Voltage", self)
        self.temps_btn = QRadioButton("Temperatures", self)
        bg.addButton(self.volts_btn)
        bg.addButton(self.temps_btn)


        layout.addWidget(self.volts_btn, 0, 0)
        layout.addWidget(self.temps_btn, 0, 1)






class BMSOverview(QWidget):

    current: QLabel
    main_voltage: QLabel
    aux_voltage: QLabel

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        # self.setMaximumWidth()
        layout = QGridLayout()
        layout.setRowStretch(0, 80)
        layout.setRowStretch(1, 20)

        number_font = QFontDatabase.systemFont(QFontDatabase.SystemFont.FixedFont)
        number_font.setPointSize(18)
        hcenter = Qt.AlignmentFlag.AlignHCenter

        self.main_voltage = QLabel("0.000", self)
        self.main_voltage.setAlignment(hcenter)
        self.main_voltage.setFont(number_font)
        layout.addWidget(self.main_voltage, 0, 0)

        main_v_label = QLabel("Main Voltage", self)
        main_v_label.setAlignment(hcenter)
        layout.addWidget(main_v_label, 1, 0)


        self.aux_voltage = QLabel("0.000", self)
        self.aux_voltage.setAlignment(hcenter)
        self.aux_voltage.setFont(number_font)
        layout.addWidget(self.aux_voltage, 0, 1)

        aux_v_label = QLabel("Aux Voltage", self)
        aux_v_label.setAlignment(hcenter)
        layout.addWidget(aux_v_label, 1, 1)

        self.current = QLabel("0.000", self)
        self.current.setAlignment(hcenter)
        self.current.setFont(number_font)
        layout.addWidget(self.current, 0, 2)

        current_label = QLabel("Battery Current", self)
        current_label.setAlignment(hcenter)
        layout.addWidget(current_label, 1, 2)

        # now add widgets that display the numeric values.
        # then make slots that take floats and display them.
        self.setLayout(layout)


    @Slot(float)
    def update_main_v(self, value: float):
        self.main_voltage.setText(f"{value:.2f}")

    @Slot(float)
    def set_aux_v(self, value:float):
        self.aux_voltage.setText(f"{value:.3f}")

    @Slot(float)
    def set_current(self, value: float):
        self.current.setText(f"{value:.3f}")



class BMSStatus(QWidget):

    contactor_items: Dict[str, QLabel] = dict()
    "A mapping of string names to the label, used to set open/closed"

    def __init__(self, parent: QWidget | None = None, contactors: list[str] = []):
        super().__init__(parent)

        layout = QVBoxLayout(self)

        self.contactors_grp = QGroupBox("Contactor State", self)
        contactor_layout = QGridLayout()
        self.contactors_grp.setLayout(contactor_layout)
        layout.addWidget(self.contactors_grp)

        for c in contactors:
            label = QLabel(c, self)
            
            self.contactor_items[c] = label
            contactor_layout.addWidget(label)




class BMSPlotsWidget(QWidget):
    pass


