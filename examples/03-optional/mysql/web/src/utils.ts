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
    if (value instanceof Uint8Array) {
        return JSON.stringify(`${value.length} bytes`);
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

const TITLES = [
    "Optional fields in practice", "Why NULL is not zero", "Writing a schema DSL",
    "SQLite in production", "Protobuf field presence", "Ten codegen mistakes",
];

const AUTHORS = ["john", "anna", "peter", "kate"];

const SUBTITLES = [
    "A short field guide", "Notes from the trenches", "With runnable examples",
];

export function randomTitle(): string {
    return TITLES[Math.floor(Math.random() * TITLES.length)];
}

export function randomAuthor(): string {
    return AUTHORS[Math.floor(Math.random() * AUTHORS.length)];
}

export function randomSubtitle(): string {
    return SUBTITLES[Math.floor(Math.random() * SUBTITLES.length)];
}

// slug is unique, so add a random suffix
export function slugFor(title: string): string {
    const base = title.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/(^-|-$)/g, "");
    return `${base}-${createHash(4)}`;
}

// fake bytes for cover_image
export function fakeCoverImage(): Uint8Array {
    const bytes = new Uint8Array(16);
    for (let i = 0; i < bytes.length; i++) {
        bytes[i] = Math.floor(Math.random() * 256);
    }
    return bytes;
}

export function daysAgo(days: number): Date {
    return new Date(Date.now() - days * 24 * 60 * 60 * 1000);
}

export function numberInput(id: string): number {
    const input = document.getElementById(id) as HTMLInputElement;
    return parseInt(input.value);
}

export function textInput(id: string): string {
    const input = document.getElementById(id) as HTMLInputElement;
    return input.value;
}

export function checkboxInput(id: string): boolean {
    const input = document.getElementById(id) as HTMLInputElement;
    return input.checked;
}
