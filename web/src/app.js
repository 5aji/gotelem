import openmct from 'openmct';
import PhoebusPlugin from "./phoebusPlugin";
openmct.setAssetPath('openmct');
openmct.install(openmct.plugins.LocalStorage());
openmct.install(openmct.plugins.MyItems());
openmct.install(openmct.plugins.Timeline());
openmct.install(openmct.plugins.UTCTimeSystem());
openmct.install(openmct.plugins.Clock({ enableClockIndicator: true }));
openmct.install(openmct.plugins.Timer());
openmct.install(openmct.plugins.Timelist());
openmct.time.clock('local', {start: -5 * 60 * 1000, end: 0});
openmct.time.timeSystem('utc');
openmct.install(openmct.plugins.Espresso());

      openmct.install(
        openmct.plugins.Conductor({
          menuOptions: [
            {
              name: 'Fixed',
              timeSystem: 'utc',
              bounds: {
                start: Date.now() - 30000000,
                end: Date.now()
              },

            },
            {
              name: 'Realtime',
              timeSystem: 'utc',
              clock: 'local',
              clockOffsets: {
                start: -30000000,
                end: 30000
              },
             
            
            }
          ]
        })
      );



if (process.env.BASE_URL) {
    console.log("got a thing")
    console.log(process.env.BASE_URL)
}

var schemaCached = null;
function getSchema() {
    if (schemaCached === null) {
        return fetch(`${process.env.BASE_URL}/api/v1/schema`).then((resp) => {
            const res = resp.json()
            console.log("got schema, caching", res);
            schemaCached = res
            return res
        })
    }
    return Promise.resolve(schemaCached)
}

const objectProvider = {
    get: function (id) {
    return getSchema().then((schema) => {
        if (id.key === "car") {
            const comp = schema.packets.map((x) => {
                return {
                    key: x.name,
                    namespace: "umnsvp"
                }
            })
            return {
                identifier: id,
                name: "the solar car",
                type: 'folder',
                location: 'ROOT',
                composition: comp
            }
        }
        var pkt = schema.packets.find((x) => x.name === id.key)
        if (pkt) {
            // if the key matches one of the packet names,
            // we know it's a field.
            const comp = pkt.data.map((field) => {
                return {
                    // we have to do this since
                    // we can't get the packet name otherwise.
                    key: `${pkt.name}.${field.name}`,
                    namespace: "umnsvp"
                }
            })
            return {
                identifier: id,
                name: pkt.name,
                type: 'folder',
                composition: comp
            }
        }
        // at this point it's definitely a field aka umnsvp-datum
        var [pktName, fieldName] = id.key.split('.')
        return {
            identifier: id,
            name: fieldName,
            type: 'umnsvp-datum',
            telemetry: {
                values: [
                    {
                        key: "value",
                        source: "val",
                        name: "Value",
                        "format": "float",
                        hints: {
                            range: 1
                        }
                    },
                    {
                        key: "utc",
                        source: "ts",
                        name: "Timestamp",
                        format: "utc",
                        hints: {
                            domain: 1
                        }

                    }
                ]
            }
        }

    })
    }
}

const TelemHistoryProvider = {
    supportsRequest: function (dObj) {
        return dObj.type === 'umnsvp-datum'
    },
    request: function (dObj, opt) {
        var [pktName, fieldName] = dObj.identifier.key.split('.')
        var url = `${process.env.BASE_URL}/api/v1/packets/${pktName}/${fieldName}?`
        var params = new URLSearchParams({
            start: new Date(opt.start).toISOString(),
            end: new Date(opt.end).toISOString(),
        }) 
        return fetch(url + params).then((resp) => {
            return resp.json()
        })

    }
}


function GotelemPlugin() {
    return function install(openmct) {

        openmct.types.addType('umnsvp-datum', {
            name: "UMN SVP Data Field",
            description: "A data field of a packet from the car",
            creatable: false,
            cssClass: "icon-telemetry"
        })
        openmct.objects.addRoot({
            namespace: "umnsvp",
            key: 'car'
        }, openmct.priority.HIGH)
        openmct.objects.addProvider('umnsvp', objectProvider);
        openmct.telemetry.addProvider(TelemHistoryProvider)
    }
}

openmct.install(GotelemPlugin())

openmct.start();
