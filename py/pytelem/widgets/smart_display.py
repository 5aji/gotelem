# A simple display for numbers with optional trend_data line, histogram, min/max, and rolling average.
from PySide6.QtCore import Qt, Slot, QSize
from PySide6.QtGui import QAction, QFontDatabase
from PySide6.QtWidgets import (
    QWidget, QVBoxLayout, QLabel, QSizePolicy, QGridLayout
)
import numpy as np

import pyqtgraph as pg
from typing import Optional, List


class _StatsDisplay(QWidget):
    """Helper Widget for the stats display."""

    def __init__(self, parent=None):
        super().__init__(parent)
        # create grid array, minimum size vertically.
        layout = QGridLayout(self)
        self.setSizePolicy(QSizePolicy.Preferred, QSizePolicy.Fixed)

    @Slot(float, float, float)
    def update_values(self, new_min: float, new_avg: float, new_max: float):



class SmartDisplay(QWidget):
    """A simple numeric display with optional statistics, trends, and histogram"""

    value: float = 0.0
    min: float = -float("inf")
    max: float = float("inf")
    avg: float = 0.0
    trend_data: List[float] = []
    histogram_data: List[float] = []

    # TODO: settable sample count for histogram/trend in right click menu

    def __init__(self, parent=None, title: str = None, initial_value: float = None, unit_suffix=None,
                 show_histogram=False, show_trendline: bool = False, show_stats=False,
                 histogram_samples=100, trend_samples=30):
        super().__init__(parent)
        self.trend_samples = trend_samples
        self.histogram_samples = histogram_samples
        layout = QVBoxLayout(self)
        if title is not None:
            self.title = title
            # create the title label
            self.title_widget = QLabel(title, self)
            self.title_widget.setAlignment(Qt.AlignmentFlag.AlignHCenter)
            self.title_widget.setSizePolicy(QSizePolicy.Preferred, QSizePolicy.Fixed)
            layout.addWidget(self.title_widget)

        number_font = QFontDatabase.systemFont(QFontDatabase.SystemFont.FixedFont)
        number_font.setPointSize(18)
        self.value = initial_value
        self.suffix = unit_suffix or ""
        self.value_widget = QLabel(f"{self.value}{self.suffix}", self)
        self.value_widget.setAlignment(Qt.AlignmentFlag.AlignHCenter)
        self.value_widget.setFont(number_font)
        layout.addWidget(self.value_widget)

        # histogram widget
        self.histogram_widget = pg.PlotWidget(self, title="Histogram")
        self.histogram_widget.enableAutoRange()
        self.histogram_widget.setVisible(False)
        self.histogram_graph = pg.PlotDataItem()
        self.histogram_widget.addItem(self.histogram_graph)

        layout.addWidget(self.histogram_widget)

        # stats display

        # trendline display
        self.trendline_widget = pg.PlotWidget(self, title="Trend")
        self.trendline_widget.enableAutoRange()
        self.trendline_widget.setVisible(False)
        self.trendline_data = pg.PlotDataItem()
        self.trendline_widget.addItem(self.trendline_data)

        layout.addWidget(self.trendline_widget)
        toggle_histogram = QAction("Show Histogram", self, checkable=True)
        toggle_histogram.toggled.connect(self._toggle_histogram)
        self.addAction(toggle_histogram)

        toggle_trendline = QAction("Show Trendline", self, checkable=True)
        toggle_trendline.toggled.connect(self._toggle_trendline)
        self.addAction(toggle_trendline)

        reset_stats = QAction("Reset Data", self)
        reset_stats.triggered.connect(self.reset_data)
        self.addAction(reset_stats)

        # use the QWidget Actions list as the right click context menu. This is inherited by children.
        self.setContextMenuPolicy(Qt.ActionsContextMenu)


    def _toggle_histogram(self):
        self.histogram_widget.setVisible(not self.histogram_widget.isVisible())

    def _toggle_trendline(self):
        self.trendline_widget.setVisible(not self.trendline_widget.isVisible())

    def _update_view(self):
        self.trendline_data.setData(self.trend_data)
        self.value_widget.setText(f"{self.value:4g}{self.suffix}")
        if self.histogram_widget.isVisible():
            hist, bins = np.histogram(self.histogram_data)
            self.histogram_graph.setData(bins, hist, stepMode="center")

    @Slot(float)
    def update_value(self, value: float):
        """Update the value displayed and associated stats."""
        self.value = value

        # update stats.
        if self.value > self.max:
            self.max = self.value
        if self.value < self.min:
            self.min = self.value

        # update trend_data data.
        self.trend_data.append(value)
        if len(self.trend_data) > self.trend_samples:
            self.trend_data.pop(0)

        # update histogram
        self.histogram_data.append(value)
        if len(self.histogram_data) > self.histogram_samples:
            self.histogram_data.pop(0)

        # update average
        # noinspection PyTypeChecker
        self.avg = np.cumsum(self.trend_data) / len(self.trend_data)

        # re-render data.
        self._update_view()

    @Slot()
    def reset_data(self):
        """Resets the existing data (trendline, stats, histogram)"""
        self.max = float("inf")
        self.min = -float("inf")
        self.trend_data = []
        self.histogram_data = []
