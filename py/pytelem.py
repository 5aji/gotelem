import time
import numpy as np

from imgui_bundle import implot, imgui_knobs, imgui, immapp, hello_imgui
import aiohttp
import orjson

# Fill x and y whose plot is a heart
vals = np.arange(0, np.pi * 2, 0.01)
x = np.power(np.sin(vals), 3) * 16
y = 13 * np.cos(vals) - 5 * np.cos(2 * vals) - 2 * np.cos(3 * vals) - np.cos(4 * vals)
# Heart pulse rate and time tracking
phase = 0
t0 = time.time() + 0.2
heart_pulse_rate = 80


class PacketState:
    """PacketState is the state representation for a packet. It contains metadata about the packet
    as well as a description of the packet fields. Also contains a buffer.
    """

    def render_tree(self):
        """Render the Tree view entry for the packet. Only called if the packet is shown."""
        pass

    def render_graphs(self):
        pass

    def __init__(self, name: str, description: str | None = None):
        self.name = name
        self.description = description

        # take the data fragment and create internal data representing it.


boards = {
    "bms": {
        "bms_measurement": {
            "description": "Voltages for main battery and aux pack",
            "id": 0x10,
            "data": {
                "battery_voltage": 127.34,
                "aux_voltage": 23.456,
                "current": 1.23,
            },
        },
        "battery_status": {
            "description": "Status bits for the battery",
            "id": 0x11,
            "data": {
                "battery_state": {
                    "startup": True,
                    "precharge": False,
                    "discharging": False,
                    "lv_only": False,
                    "charging": False,
                    "wall_charging": False,
                    "killed": False,
                },  # repeat for rest fo fields
            },
        },
    }
}


def gui():
    global heart_pulse_rate, phase, t0, x, y
    # Make sure that the animation is smooth
    hello_imgui.get_runner_params().fps_idling.enable_idling = False

    t = time.time()
    phase += (t - t0) * heart_pulse_rate / (np.pi * 2)
    k = 0.8 + 0.1 * np.cos(phase)
    t0 = t

    imgui.show_demo_window()
    main_window_flags: imgui.WindowFlags = imgui.WindowFlags_.no_collapse.value
    imgui.begin("my application", p_open=None, flags=main_window_flags)

    imgui.text("Bloat free code")
    if implot.begin_plot("Heart", immapp.em_to_vec2(21, 21)):
        implot.plot_line("", x * k, y * k)
        implot.end_plot()

    for board_name, board_packets in boards.items():
        if imgui.tree_node(board_name):
            for packet_name in board_packets:
                if imgui.tree_node(packet_name):
                    # display description if hovered
                    pkt = board_packets[packet_name]
                    if imgui.is_item_hovered():
                        imgui.set_tooltip(pkt["description"])
                    imgui.text(f"0x{pkt['id']:03X}")
                    imgui.tree_pop()
            imgui.tree_pop()
    imgui.end()  # my application

    _, heart_pulse_rate = imgui_knobs.knob("Pulse", heart_pulse_rate, 30, 180)


# class State:
#     def __init__(self):
#
#     def gui(self):

if __name__ == "__main__":
    immapp.run(
        gui,
        window_size=(300, 450),
        window_title="Hello!",
        with_implot=True,
        fps_idle=0,
    )  # type: ignore
