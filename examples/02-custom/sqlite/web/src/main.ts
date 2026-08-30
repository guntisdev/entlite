import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import { ReadingService, SensorService } from "../../ent/gen/ts/schema_pb.js";
import type {
    CreateReadingRequest,
    CreateSensorRequest,
    ListReadingFilterBySensorIdRecordedAtFlaggedRequest,
    ListSensorFilterByLabelKindActiveRequest,
    Sensor,
    UpdateSensorRequest,
} from "../../ent/gen/ts/schema_pb.js";
// from custom proto
import { SensorAnalyticsService } from "../../ent/gen/ts/custom_pb.js";
import type {
    GetSensorReadingStatsRequest,
    ListSensorsWithLatestReadingRequest,
    PruneReadingsRequest,
} from "../../ent/gen/ts/custom_pb.js";
import {
    bigIntInput,
    checkboxInput,
    createHash,
    daysAgo,
    numberInput,
    randomKind,
    randomLocation,
    randomValue,
    textInput,
    toString,
    unitFor,
} from "./utils.js";

type StrictMessageInput<T extends { $typeName: string; $unknown?: unknown }> = Omit<T, "$typeName" | "$unknown">;

const transport = createConnectTransport({
    baseUrl: "http://localhost:8080",
});

const sensorClient = createClient(SensorService, transport);
const readingClient = createClient(ReadingService, transport);
const analyticsClient = createClient(SensorAnalyticsService, transport);

function log(message: string, data?: any) {
    console.log(message, data);
    const output = document.getElementById("output")!;
    const line = document.createElement("div");
    line.textContent = data !== undefined ? `${message} ${toString(data)}` : message;
    output.appendChild(line);
    output.scrollTop = output.scrollHeight;
}

function describeSensor(sensor: Sensor): string {
    const latest = sensor.latestValue !== undefined ? ` latest=${sensor.latestValue}` : "";
    return `ID: ${sensor.ID} ${sensor.code} ${sensor.kind} ${sensor.label}${latest}`;
}

// --- SensorService: generated from the DSL ---------------------------------

function createSensor() {
    const kind = randomKind();
    const code = `${kind.slice(0, 4).toUpperCase()}-${createHash(4).toUpperCase()}`;
    log(`Creating ${kind} sensor ${code}...`);
    const request: StrictMessageInput<CreateSensorRequest> = {
        code: code,
        label: `${kind[0].toUpperCase()}${kind.slice(1)} probe ${createHash(3)}`,
        kind: kind,
        unit: unitFor(kind),
        location: randomLocation(),
        sampleRateMs: 1000 * Math.ceil(Math.random() * 5),
        installedAt: timestampFromDate(daysAgo(Math.ceil(Math.random() * 90))),
        // active and firmware are left unset on purpose: the DSL defaults
        // (true / "1.0.0") are applied by the generated create query
    };
    sensorClient.create(request)
        .then((response) => {
            log("✓ Sensor created:", response);
        })
        .catch((error) => {
            log("✗ Error creating sensor:", error);
        });
}

function createInvalidSensor() {
    log("Creating sensor with kind 'plasma' (rejected by logic.IsKnownSensorKind)...");
    const request: StrictMessageInput<CreateSensorRequest> = {
        code: `BAD-${createHash(4).toUpperCase()}`,
        label: "Invalid kind probe",
        kind: "plasma",
        unit: "kelvin",
        installedAt: timestampFromDate(new Date()),
    };
    sensorClient.create(request)
        .then((response) => {
            log("✓ Sensor created (unexpected):", response);
        })
        .catch((error) => {
            log("✗ Rejected by the Validate() interceptor:", error);
        });
}

function getSensorByID() {
    const id = numberInput("getSensorId");
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Getting sensor ${id}...`);
    sensorClient.getByID({ ID: id })
        .then((response) => {
            log("✓ Sensor retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting sensor:", error);
        });
}

function getSensorByCode() {
    const code = textInput("getSensorCode");
    if (code === "") {
        log("✗ Invalid sensor code");
        return;
    }
    log(`Getting sensor by code ${code}...`);
    sensorClient.getByCode({ code: code })
        .then((response) => {
            log("✓ Sensor retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting sensor by code:", error);
        });
}

function updateSensor() {
    const id = numberInput("updateSensorId");
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Updating sensor ${id}...`);
    // Update needs the whole entity, so read it first and change a few fields
    sensorClient.getByID({ ID: id })
        .then((sensor) => {
            const request: StrictMessageInput<UpdateSensorRequest> = {
                ID: sensor.ID,
                code: sensor.code,
                label: `Updated ${createHash(3)}`,
                kind: sensor.kind,
                unit: sensor.unit,
                location: randomLocation(),
                active: !sensor.active,
                firmware: `1.0.${Math.ceil(Math.random() * 9)}`,
                sampleRateMs: 1000 * Math.ceil(Math.random() * 5),
                // installed_at and created_at are Immutable / ReadOnly in the
                // DSL, so they are not part of the update request at all
            };
            return sensorClient.update(request);
        })
        .then((response) => {
            log("✓ Sensor updated:", response);
        })
        .catch((error) => {
            log("✗ Error updating sensor:", error);
        });
}

function deleteSensor() {
    const id = numberInput("deleteSensorId");
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Deleting sensor ${id}...`);
    sensorClient.delete({ ID: id })
        .then((response) => {
            log("✓ Sensor deleted:", response);
        })
        .catch((error) => {
            log("✗ Error deleting sensor:", error);
        });
}

function filterSensors() {
    const label = textInput("filterLabel");
    const kind = textInput("filterKind");
    const active = checkboxInput("filterActive");
    log(`Filtering sensors: label LIKE ${label}, kind=${kind}, active=${active}...`);
    const request: StrictMessageInput<ListSensorFilterByLabelKindActiveRequest> = {
        limit: 50,
        offset: 0,
        label: label,
        kind: kind,
        active: active,
    };
    sensorClient.filterByLabelKindActive(request)
        .then((response) => {
            log(`✓ Sensors filtered (${response.sensors.length} sensors):`);
            response.sensors.forEach((sensor) => log(describeSensor(sensor)));
        })
        .catch((error) => {
            log("✗ Error filtering sensors:", error);
        });
}

// --- ReadingService: generated from the DSL --------------------------------

function createReading() {
    const sensorId = numberInput("readingSensorId");
    if (isNaN(sensorId) || sensorId <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Creating reading for sensor ${sensorId}...`);
    // Look the sensor up first so the value matches its kind
    sensorClient.getByID({ ID: sensorId })
        .then((sensor) => {
            const request: StrictMessageInput<CreateReadingRequest> = {
                sensorId: sensor.ID,
                value: randomValue(sensor.kind as any),
                quality: 50 + Math.floor(Math.random() * 51),
                flagged: Math.random() < 0.2,
                recordedAt: timestampFromDate(daysAgo(Math.random() * 14)),
            };
            return readingClient.create(request);
        })
        .then((response) => {
            log("✓ Reading created:", response);
        })
        .catch((error) => {
            log("✗ Error creating reading:", error);
        });
}

function createInvalidReading() {
    const sensorId = numberInput("readingSensorId");
    log(`Creating reading with quality 150 (rejected by logic.IsPercentage)...`);
    const request: StrictMessageInput<CreateReadingRequest> = {
        sensorId: isNaN(sensorId) ? 1 : sensorId,
        value: 1,
        quality: 150,
        recordedAt: timestampFromDate(new Date()),
    };
    readingClient.create(request)
        .then((response) => {
            log("✓ Reading created (unexpected):", response);
        })
        .catch((error) => {
            log("✗ Rejected by the Validate() interceptor:", error);
        });
}

function getReadingByID() {
    const id = bigIntInput("getReadingId");
    if (id <= 0n) {
        log("✗ Invalid reading ID");
        return;
    }
    log(`Getting reading ${id}...`);
    readingClient.getByID({ ID: id })
        .then((response) => {
            log("✓ Reading retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting reading:", error);
        });
}

function deleteReading() {
    const id = bigIntInput("deleteReadingId");
    if (id <= 0n) {
        log("✗ Invalid reading ID");
        return;
    }
    log(`Deleting reading ${id}...`);
    readingClient.delete({ ID: id })
        .then((response) => {
            log("✓ Reading deleted:", response);
        })
        .catch((error) => {
            log("✗ Error deleting reading:", error);
        });
}

function listReadings() {
    const sensorId = numberInput("readingSensorId");
    if (isNaN(sensorId) || sensorId <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Listing readings of sensor ${sensorId}...`);
    readingClient.listBySensorId({ limit: 50, offset: 0, sensorId: sensorId })
        .then((response) => {
            log(`✓ Readings listed (${response.readings.length} readings):`);
            response.readings.forEach((reading) => {
                const recordedAt = reading.recordedAt ? timestampDate(reading.recordedAt).toISOString() : "-";
                log(`ID: ${reading.ID} value=${reading.value} quality=${reading.quality} flagged=${reading.flagged} ${recordedAt}`);
            });
        })
        .catch((error) => {
            log("✗ Error listing readings:", error);
        });
}

function filterReadings() {
    const sensorId = numberInput("readingSensorId");
    if (isNaN(sensorId) || sensorId <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Filtering readings of sensor ${sensorId} over the last 30 days...`);
    const request: StrictMessageInput<ListReadingFilterBySensorIdRecordedAtFlaggedRequest> = {
        limit: 50,
        offset: 0,
        sensorId: sensorId,
        minRecordedAt: timestampFromDate(daysAgo(30)),
        maxRecordedAt: timestampFromDate(new Date()),
        flagged: true,
    };
    readingClient.filterBySensorIdRecordedAtFlagged(request)
        .then((response) => {
            log(`✓ Readings filtered (${response.readings.length} readings):`, response);
        })
        .catch((error) => {
            // Known gap: sqlc drops both bounds of a DATETIME BETWEEN from the
            // generated params, so the query is called with too few arguments
            log("✗ Error filtering readings (known filter.Range gap):", error);
        });
}

// --- SensorAnalyticsService: hand-written custom.proto + custom.sql --------

function listWithLatestReading() {
    log("Listing active sensors with their latest reading (custom LEFT JOIN)...");
    const request: StrictMessageInput<ListSensorsWithLatestReadingRequest> = {
        limit: 50,
        offset: 0,
    };
    analyticsClient.listWithLatestReading(request)
        .then((response) => {
            log(`✓ Sensors listed (${response.items.length} sensors):`);
            response.items.forEach((item) => {
                const recordedAt = item.latestRecordedAt ? timestampDate(item.latestRecordedAt).toISOString() : "never";
                const value = item.latestValue !== undefined ? item.latestValue : "-";
                log(`${item.sensor ? describeSensor(item.sensor) : "?"} | latest ${value} at ${recordedAt}`);
            });
        })
        .catch((error) => {
            log("✗ Error listing sensors with latest reading:", error);
        });
}

function getReadingStats() {
    const sensorId = numberInput("statsSensorId");
    if (isNaN(sensorId) || sensorId <= 0) {
        log("✗ Invalid sensor ID");
        return;
    }
    log(`Getting reading stats of sensor ${sensorId} over the last 30 days...`);
    const request: StrictMessageInput<GetSensorReadingStatsRequest> = {
        sensorId: sensorId,
        fromTs: timestampFromDate(daysAgo(30)),
        toTs: timestampFromDate(new Date()),
    };
    analyticsClient.getReadingStats(request)
        .then((response) => {
            log(`✓ Stats: count=${response.readingCount} avg=${response.avgValue} min=${response.minValue} max=${response.maxValue}`);
        })
        .catch((error) => {
            log("✗ Error getting reading stats:", error);
        });
}

function pruneReadings() {
    const days = numberInput("pruneDays");
    if (isNaN(days) || days < 0) {
        log("✗ Invalid number of days");
        return;
    }
    log(`Pruning readings older than ${days} days...`);
    const request: StrictMessageInput<PruneReadingsRequest> = {
        cutoff: timestampFromDate(daysAgo(days)),
    };
    analyticsClient.pruneReadings(request)
        .then((response) => {
            log(`✓ Readings pruned: ${response.deleted} deleted`);
        })
        .catch((error) => {
            log("✗ Error pruning readings:", error);
        });
}

document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("createSensorBtn")!.addEventListener("click", createSensor);
    document.getElementById("createInvalidSensorBtn")!.addEventListener("click", createInvalidSensor);
    document.getElementById("getSensorBtn")!.addEventListener("click", getSensorByID);
    document.getElementById("getSensorCodeBtn")!.addEventListener("click", getSensorByCode);
    document.getElementById("updateSensorBtn")!.addEventListener("click", updateSensor);
    document.getElementById("deleteSensorBtn")!.addEventListener("click", deleteSensor);
    document.getElementById("filterSensorBtn")!.addEventListener("click", filterSensors);

    document.getElementById("createReadingBtn")!.addEventListener("click", createReading);
    document.getElementById("createInvalidReadingBtn")!.addEventListener("click", createInvalidReading);
    document.getElementById("listReadingBtn")!.addEventListener("click", listReadings);
    document.getElementById("filterReadingBtn")!.addEventListener("click", filterReadings);
    document.getElementById("getReadingBtn")!.addEventListener("click", getReadingByID);
    document.getElementById("deleteReadingBtn")!.addEventListener("click", deleteReading);

    document.getElementById("latestBtn")!.addEventListener("click", listWithLatestReading);
    document.getElementById("statsBtn")!.addEventListener("click", getReadingStats);
    document.getElementById("pruneBtn")!.addEventListener("click", pruneReadings);

    document.getElementById("clearBtn")!.addEventListener("click", () => {
        document.getElementById("output")!.innerHTML = "";
    });

    log("Entlite Custom Demo Ready!");
    log("SensorService and ReadingService come from the DSL, SensorAnalyticsService from custom.proto + custom.sql");
});
