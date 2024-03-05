import openmct from 'openmct';
import PhoebusPlugin from "./phoebusPlugin";
openmct.setAssetPath('openmct');
openmct.install(openmct.plugins.LocalStorage());
openmct.install(openmct.plugins.MyItems());
openmct.install(openmct.plugins.UTCTimeSystem());
openmct.time.clock('local', {start: -5 * 60 * 1000, end: 0});
openmct.time.timeSystem('utc');
openmct.install(openmct.plugins.Espresso());

if (process.env.BASE_URL) {
    console.log("got a thing")
    console.log(process.env.BASE_URL)
}
function GotelemPlugin() {

}

openmct.start();
