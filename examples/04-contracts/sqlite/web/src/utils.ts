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

const PLAYERS = ["capablanca", "tal", "polgar", "carlsen", "gukesh"];

const OPENINGS = [
    "Sicilian Defence", "Queen's Gambit", "London System", "Caro-Kann", "King's Indian",
];

const RESULTS = ["1-0", "0-1", "1/2-1/2"];

export function randomPlayer(exclude?: string): string {
    const pool = exclude ? PLAYERS.filter((name) => name !== exclude) : PLAYERS;
    return pool[Math.floor(Math.random() * pool.length)];
}

export function randomOpening(): string {
    return OPENINGS[Math.floor(Math.random() * OPENINGS.length)];
}

export function randomResult(): string {
    return RESULTS[Math.floor(Math.random() * RESULTS.length)];
}

export function randomMoves(): number {
    return 20 + Math.floor(Math.random() * 60);
}

export function daysAgo(days: number): Date {
    return new Date(Date.now() - days * 24 * 60 * 60 * 1000);
}

export function numberInput(id: string): number {
    const input = document.getElementById(id) as HTMLInputElement;
    return parseInt(input.value);
}
