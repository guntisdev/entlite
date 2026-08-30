export function toString(value: unknown): string {
    if (value === undefined) {
        return "null";
    }
    if (typeof value === "bigint") {
        return value.toString() + "n";
    }
    if (value === null) {
        return "null";
    }
    if (value instanceof Error) {
        const errorObj: any = {
            message: value.message,
            name: value.name,
        };
        for (const [key, val] of Object.entries(value)) {
            const stringResult = toString(val);
            try {
                errorObj[key] = JSON.parse(stringResult);
            } catch {
                errorObj[key] = stringResult;
            }
        }
        return JSON.stringify(errorObj);
    }
    if (typeof value === "object") {
        const converted: any = Array.isArray(value) ? [] : {};
        for (const [key, val] of Object.entries(value)) {
            const stringResult = toString(val);
            try {
                converted[Array.isArray(value) ? parseInt(key) : key] = JSON.parse(stringResult);
            } catch {
                converted[Array.isArray(value) ? parseInt(key) : key] = stringResult;
            }
        }
        return JSON.stringify(converted);
    }
    const result = JSON.stringify(value);
    return result !== undefined ? result : "null";
}

export function createHash(length: number = 4): string {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

// The kinds accepted by logic.IsKnownSensorKind - anything else is rejected by
// the generated Validate() interceptor before it reaches the database
export const SENSOR_KINDS = ["temperature", "humidity", "pressure", "motion"] as const;

export type SensorKind = typeof SENSOR_KINDS[number];

const UNITS: Record<SensorKind, string> = {
    temperature: "celsius",
    humidity: "percent",
    pressure: "hectopascal",
    motion: "events",
};

const LOCATIONS = [
    "Warehouse A", "Warehouse B", "Cold room", "Roof", "Boiler room",
    "Server rack 3", "Loading bay", "Greenhouse",
];

export function randomKind(): SensorKind {
    return SENSOR_KINDS[Math.floor(Math.random() * SENSOR_KINDS.length)];
}

export function unitFor(kind: SensorKind): string {
    return UNITS[kind];
}

export function randomLocation(): string {
    return LOCATIONS[Math.floor(Math.random() * LOCATIONS.length)];
}

// Plausible measurement for the given kind, so the analytics numbers look real
export function randomValue(kind: SensorKind): number {
    switch (kind) {
        case "temperature":
            return round(-10 + Math.random() * 45);
        case "humidity":
            return round(Math.random() * 100);
        case "pressure":
            return round(950 + Math.random() * 100);
        case "motion":
            return Math.floor(Math.random() * 20);
    }
}

export function round(n: number, decimals: number = 2): number {
    const factor = 10 ** decimals;
    return Math.round(n * factor) / factor;
}

export function daysAgo(days: number): Date {
    return new Date(Date.now() - days * 24 * 60 * 60 * 1000);
}

export function numberInput(id: string): number {
    const input = document.getElementById(id) as HTMLInputElement;
    return parseInt(input.value);
}

// int64 fields are bigint in TS, so their ID inputs parse to bigint, not number
export function bigIntInput(id: string): bigint {
    const input = document.getElementById(id) as HTMLInputElement;
    try {
        return BigInt(input.value);
    } catch {
        return 0n;
    }
}

export function textInput(id: string): string {
    const input = document.getElementById(id) as HTMLInputElement;
    return input.value;
}

export function checkboxInput(id: string): boolean {
    const input = document.getElementById(id) as HTMLInputElement;
    return input.checked;
}
