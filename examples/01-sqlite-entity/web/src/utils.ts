export function toString(value: unknown): string {
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
            errorObj[key] = JSON.parse(toString(val));
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
    return JSON.stringify(value);
}

export function createHash(length: number = 4): string {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

const NAMES = [
  'James', 'John', 'Robert', 'Michael', 'William', 'David', 'Richard', 'Joseph',
  'Thomas', 'Charles', 'Mary', 'Patricia', 'Jennifer', 'Linda', 'Elizabeth',
  'Barbara', 'Susan', 'Jessica', 'Sarah', 'Karen', 'Emma', 'Olivia', 'Ava',
  'Isabella', 'Sophia', 'Mia', 'Charlotte', 'Amelia', 'Harper', 'Evelyn',
  'Daniel', 'Matthew', 'Anthony', 'Mark', 'Donald', 'Steven', 'Paul', 'Andrew',
  'Joshua', 'Kenneth', 'Kevin', 'Brian', 'George', 'Timothy', 'Ronald', 'Edward',
  'Jason', 'Jeffrey', 'Ryan', 'Jacob', 'Gary', 'Nicholas', 'Eric', 'Jonathan',
  'Stephen', 'Larry', 'Justin', 'Scott', 'Brandon', 'Benjamin', 'Samuel', 'Frank',
  'Smith', 'Johnson', 'Williams', 'Brown', 'Jones', 'Garcia', 'Miller', 'Davis',
  'Rodriguez', 'Martinez', 'Hernandez', 'Lopez', 'Gonzalez', 'Wilson', 'Anderson',
  'Moore', 'Jackson', 'Martin', 'Lee', 'Perez', 'Thompson', 'White', 'Harris',
  'Sanchez', 'Clark', 'Ramirez', 'Lewis', 'Robinson', 'Walker', 'Young', 'Allen',
  'King', 'Wright', 'Scott', 'Torres', 'Nguyen', 'Hill', 'Flores', 'Green',
  'Adams', 'Nelson', 'Baker', 'Hall', 'Rivera', 'Campbell', 'Mitchell', 'Carter',
  'Roberts', 'Gomez', 'Phillips', 'Evans', 'Turner', 'Diaz', 'Parker', 'Cruz',
  'Edwards', 'Collins', 'Reyes', 'Stewart', 'Morris', 'Morales', 'Murphy', 'Cook',
  'Rogers', 'Morgan', 'Peterson', 'Cooper', 'Reed', 'Bailey', 'Bell', 'Howard'
];

export function randomName(): string {
  return NAMES[Math.floor(Math.random() * NAMES.length)];
}

export function randomFullName(): string {
  const firstName = randomName();
  const lastName = randomName();
  
  // 30% chance of having a middle name
  if (Math.random() < 0.3) {
    const middleName = randomName();
    return `${firstName} ${middleName} ${lastName}`;
  }
  
  return `${firstName} ${lastName}`;
}
