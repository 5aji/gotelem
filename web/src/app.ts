import openmct from "openmct";

//@ts-expect-error openmct
openmct.setAssetPath('openmct');
//@ts-expect-error openmct
openmct.install(openmct.plugins.LocalStorage());
//@ts-expect-error openmct
openmct.install(openmct.plugins.MyItems());
//@ts-expect-error openmct
openmct.install(openmct.plugins.Timeline());
//@ts-expect-error openmct
openmct.install(openmct.plugins.UTCTimeSystem());
//@ts-expect-error openmct
openmct.install(openmct.plugins.Clock({ enableClockIndicator: true }));
//@ts-expect-error openmct
openmct.install(openmct.plugins.Timer());
//@ts-expect-error openmct
openmct.install(openmct.plugins.Timelist());
//@ts-expect-error openmct
openmct.install(openmct.plugins.Hyperlink())
//@ts-expect-error openmct
openmct.install(openmct.plugins.Notebook())
//@ts-expect-error openmct
openmct.install(openmct.plugins.BarChart())
//@ts-expect-error openmct
openmct.install(openmct.plugins.ScatterPlot())
//@ts-expect-error openmct
openmct.install(openmct.plugins.SummaryWidget())
//@ts-expect-error openmct
openmct.install(openmct.plugins.LADTable());
openmct.time.clock('local', { start: -5 * 60 * 1000, end: 0 });
//@ts-expect-error openmct
openmct.time.timeSystem('utc');
//@ts-expect-error openmct
openmct.install(openmct.plugins.Espresso());

openmct.install(
//@ts-expect-error openmct
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
interface SkylabField {
    name: string
    type: string
    units?: string
    conversion?: number
    bits?: {
        name: string
    }
}
interface SkylabPacket {
    name: string
    description?: string
    id: number
    endian?: string
    is_extended: boolean
    repeat?: number
    offset?: number
    data: [SkylabField]
}

interface SkylabBoard {
    name: string
    transmit: [string]
    receive: [string]
}

interface SkylabSchema {
    packets: [SkylabPacket]
    boards: [SkylabBoard]
}

let schemaCached = null;
function getSchema(): Promise<SkylabSchema> {
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
            const pkt = schema.packets.find((x) => x.name === id.key)
            if (pkt) {
                // if the key matches one of the packet names,
                // we know it's a packet.

                // construct a list of fields for this packet.
                const comp = pkt.data.map((field) => {
                    if (field.type === "bitfield") {
                        // 

                    }
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
            const [pktName, fieldName] = id.key.split('.')
            return {
                identifier: id,
                name: fieldName,
                type: 'umnsvp-datum',
                conversion: schema.packets.find((x) => x.name === pktName).data.find((f) => f.name === fieldName).conversion,
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


interface Datum {
    ts: number
    val: number
}
const TelemHistoryProvider = {
    supportsRequest: function (dObj) {
        return dObj.type === 'umnsvp-datum'
    },
    request: function (dObj, opt) {
        const [pktName, fieldName] = dObj.identifier.key.split('.')
        const url = `${process.env.BASE_URL}/api/v1/packets/${pktName}/${fieldName}?`
        const params = new URLSearchParams({
            start: new Date(opt.start).toISOString(),
            end: new Date(opt.end).toISOString(),
        })
        console.log((opt.end - opt.start) / opt.size)
        return fetch(url + params).then((resp) => {
            resp.json().then((result: [Datum]) => {

                if (dObj.conversion && dObj.conversion != 0) {
                    // apply conversion
                    result.map((dat) => {
                        dat.val = dat.val * dObj.conversion
                        return dat
                    })
                }
                return result
            })
        })

    }
}

interface PacketData {
    ts: number
    name: string
    data: object
}

function TelemRealtimeProvider() {
    return function (openmct: openmct.OpenMCT) {

        const simpleIndicator = openmct.indicators.simpleIndicator();
        openmct.indicators.add(simpleIndicator);
        simpleIndicator.text("0 Listeners")
        const url = `${process.env.BASE_URL.replace(/^http/, 'ws')}/api/v1/packets/subscribe?`
        // we put our websocket connection here.
        let connection = new WebSocket(url)
        // connections contains name: callback mapping
        const callbacks = {}

        const conversions: Map<string, number> = new Map()
        // names contains a set of *packet names*
        const names = new Set()

        function handleMessage(event) {
            const data: PacketData = JSON.parse(event.data)
            for (const [key, value] of Object.entries(data.data)) { // for each of the fields in the data
                const id = `${data.name}.${key}` // if we have a matching callback for that field.
                if (id in callbacks) {
                    // we should construct a telem point and make a callback.
                    // compute if we need to scale the value.
                    callbacks[id]({
                        "ts": data.ts,
                        "val": value * conversions.get(id)
                    })
                }
            }
        }

        function updateWebsocket() {
            const params = new URLSearchParams()
            for (const name in names) {
                params.append("name", name)
            }
            connection = new WebSocket(url + params)

            connection.onmessage = handleMessage
            simpleIndicator.text(`${names.size} Listeners`)
        }

        const provider = {
            supportsSubscribe: function (dObj) {
                return dObj.type === "umnsvp-datum"
            },
            subscribe: function (dObj, callback) {
                // identifier is packetname.fieldname. we add the packet name to the set.
                const key = dObj.identifier.key
                const pktName = key.split('.')[0]
                // add our callback to the dictionary,
                // add the packet name to the set
                callbacks[key] = callback
                conversions.set(key, dObj.conversion || 1)
                names.add(pktName)
                // update the websocket URL with the new name.
                updateWebsocket()
                return function unsubscribe() {
                    // if there's no more listeners on this packet,
                    // we can remove it.
                    console.log("subscribe called %s", JSON.stringify(dObj))
                    if (!Object.keys(callbacks).some((k) => k.startsWith(pktName))) {
                        names.delete(pktName)
                        updateWebsocket()
                    }

                    delete callbacks[key]
                    conversions.delete(key)
                }
            }
        }
        openmct.telemetry.addProvider(provider)
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
        openmct.install(TelemRealtimeProvider())
    }
}

openmct.install(GotelemPlugin())

//@ts-expect-error openmct
openmct.start()
