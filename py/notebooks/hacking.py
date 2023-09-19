import orjson
import numpy as np
import pyqtgraph as pg
from pathlib import Path
from dataclasses import dataclass
import glom


# define a structure that can be used to describe what data to graph
print("hi")
@dataclass
class PlotFeature:
    """Class that represents a feature extraction"""
    pkt_name: str
    info_path: list[str]


# now make a function that takes a bunch of these and then matches the pkt_name.
# if there is a match, we must push the data.

# data format : dict[dict[list[timestamp, value]]]
# first dict is pkt_name, second dict is each variable we care about, and the
# list is a timestamp-value plot.

def rip_and_tear(fname: Path, features: list[PlotFeature]):
    data = {}
    for feat in features:
        v = {}
        for path in feat.info_path:
            v[path] = []
        data[feat.pkt_name] = v

    # now we have initialized the data structure, start parsing the file.

    with open(fname) as f:
        while line := f.readline():
            if len(line) < 3:
                continue  # kludge to skip empty lines

            j = orjson.loads(line)
            if not j['name'] in data:
                continue
            # use the glom, harry

            for path in data[j['name']].keys():
                d = glom.glom(j['data'], path)
                ts = j['ts'] - 1688756556040
                data[j['name']][path].append([ts, d])
    # TODO: numpy the last list???
    return data


if __name__ == "__main__":
    features = [
        PlotFeature("bms_measurement", ["current"]),
        PlotFeature("wsr_phase_current", ["phase_b_current"]),
        PlotFeature("wsr_motor_current_vector", ["iq"]),
        PlotFeature("wsr_motor_voltage_vector", ["vq"]),
        PlotFeature("wsr_velocity", ["motor_velocity"])
    ]
    logs_path = Path("../../logs/")
    logfile = logs_path / "RETIME_7-2-hillstart.txt"
    res = rip_and_tear(logfile, features)
    # now fuck my shit up and render some GRAPHHHSSS
    app = pg.mkQApp("i see no god up here\n OTHER THAN ME")
    win = pg.GraphicsLayoutWidget(show=True, title="boy howdy")
    prev_plot = None
    for packet_name, fields in res.items():
        win.addLabel(f"{packet_name}")
        win.nextRow()
        for field_name, field_data in fields.items():
            d = np.array(field_data)
            p = win.addPlot(title=f"{field_name}")
            if prev_plot is not None:
                p.setXLink(prev_plot)
            p.plot(d)
            prev_plot = p
            win.nextRow()
    pg.exec()




