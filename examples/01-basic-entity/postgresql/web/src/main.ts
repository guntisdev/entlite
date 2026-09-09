import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { UserService } from "../../ent/gen/ts/schema_pb.js";
import type { CreateBulkUserRow, CreateBulkUserRequest, CreateUserRequest, DeleteAllUserRequest, ListActiveRequest, ListAllUserRequest, ListUserFilterByAgeNameRequest, UpdateUserRequest } from "../../ent/gen/ts/schema_pb.js";
import { createHash, numberInput, randomFullName, randomName, textInput, toString } from "./utils.js";

type StrictMessageInput<T extends { $typeName: string; $unknown?: unknown }> = Omit<T, "$typeName" | "$unknown">;

const transport = createConnectTransport({
    baseUrl: "http://localhost:8080",
});

const client = createClient(UserService, transport);

function log(message: string, data?: any) {
    console.log(message, data);
    const output = document.getElementById("output")!;
    const line = document.createElement("div");
    line.textContent = data ? `${message} ${toString(data)}` : message;
    output.appendChild(line);
}

function createUser() {
    log("Creating user...");
    const fullName = randomFullName();
    const email = `${fullName.split(" ")[0].toLowerCase()}_${createHash()}@example.com`;
    const request: StrictMessageInput<CreateUserRequest> = {
        email: email,
        name: fullName,
        age: Math.ceil(Math.random() * 100),
        password: createHash(12),
        isAdmin: false,
        lastLoginMs: BigInt(Date.now()),
    };
    client.createUser(request)
        .then((response) => {
            log("✓ User created:", response);
        })
        .catch((error) => {
            log("✗ Error creating user:", error);
        });
}

function createBulkUsers() {
    const count = 3;
    log(`Creating ${count} users in bulk...`);
    const rows: StrictMessageInput<CreateBulkUserRow>[] = Array.from({ length: count }, () => {
        const fullName = randomFullName();
        const email = `${fullName.split(" ")[0].toLowerCase()}_${createHash()}@example.com`;
        return {
            email: email,
            name: fullName,
            age: Math.ceil(Math.random() * 100),
            password: createHash(12),
        };
    });
    const request: StrictMessageInput<CreateBulkUserRequest> = { rows };
    client.createBulkUser(request)
        .then((response) => {
            log(`✓ ${response.rows.length} users created in bulk:`, response);
        })
        .catch((error) => {
            log("✗ Error creating users in bulk:", error);
        });
}

function getUserById() {
    const idInput = document.getElementById("getId") as HTMLInputElement;
    const id = parseInt(idInput.value);
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid user ID");
        return;
    }
    log(`Getting user ${id}...`);
    client.getUserById({ id: id })
        .then((response) => {
            log("✓ User retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting user:", error);
        });
}

function listAllUsers() {
    log("Listing all users...");
    const request: StrictMessageInput<ListAllUserRequest> = {};
    client.listAllUser(request)
        .then((response) => {
            log(`✓ Users listed (${response.rows.length} users):`);
            response.rows.forEach((user, index) => {
                log(`ID: ${user.id} ${user.name} ${user.age} ${user.email}`);
            });
        })
        .catch((error) => {
            log("✗ Error listing users:", error);
        });
}

function listActiveUsers() {
    const limit = numberInput("activeLimit");
    const offset = numberInput("activeOffset");
    log(`Listing active users, limit ${limit} offset ${offset}...`);
    const request: StrictMessageInput<ListActiveRequest> = {
        limit: limit,
        offset: offset,
        isActive: true,
    };
    client.listActive(request)
        .then((response) => {
            log(`✓ Active users listed (${response.rows.length} users):`);
            response.rows.forEach((user) => {
                log(`ID: ${user.id} ${user.name} ${user.age} ${user.email}`);
            });
        })
        .catch((error) => {
            log("✗ Error listing active users:", error);
        });
}

function filterUsersByAgeName() {
    const limit = numberInput("filterLimit");
    const offset = numberInput("filterOffset");
    log(`Filtering users, limit ${limit} offset ${offset}...`);
    const request: StrictMessageInput<ListUserFilterByAgeNameRequest> = {
        limit: limit,
        offset: offset,
        minAge: numberInput("filterMinAge"),
        maxAge: numberInput("filterMaxAge"),
        name: textInput("filterName"),
    };
    client.listUserFilterByAgeName(request)
        .then((response) => {
            log(`✓ Users filtered (${response.rows.length} users):`);
            response.rows.forEach((user) => {
                log(`ID: ${user.id} ${user.name} ${user.age} ${user.email}`);
            });
        })
        .catch((error) => {
            log("✗ Error filtering users:", error);
        });
}

function updateUser() {
    const idInput = document.getElementById("updateId") as HTMLInputElement;
    const id = parseInt(idInput.value);
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid user ID");
        return;
    }
    const fullName = "Updated " + randomName();
    const email = `${fullName.split(" ")[0].toLowerCase()}_${createHash()}@example.com`;
    log(`Updating user ${id}...`);
    const request: StrictMessageInput<UpdateUserRequest> = {
        id: id,
        email: email,
        name: fullName,
        age: Math.ceil(Math.random() * 100),
        isAdmin: true,
        lastLoginMs: BigInt(Date.now()),
    };
    client.updateUser(request)
        .then((response) => {
            log("✓ User updated:", response);
        })
        .catch((error) => {
            log("✗ Error updating user:", error);
        });
}

function deleteUser() {
    const idInput = document.getElementById("deleteId") as HTMLInputElement;
    const id = parseInt(idInput.value);
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid user ID");
        return;
    }
    log(`Deleting user ${id}...`);
    client.deleteUser({ id: id })
        .then((response) => {
            log("✓ User deleted:", response);
        })
        .catch((error) => {
            log("✗ Error deleting user:", error);
        });
}

function deleteAllUsers() {
    log("Deleting all users...");
    const request: StrictMessageInput<DeleteAllUserRequest> = {};
    client.deleteAllUser(request)
        .then((response) => {
            log("✓ All users deleted:", response);
        })
        .catch((error) => {
            log("✗ Error deleting all users:", error);
        });
}

document.addEventListener("DOMContentLoaded", () => {
    console.log("DOMContentLoaded", document.getElementById("createBtn"));
    document.getElementById("createBtn")!.addEventListener("click", createUser);
    document.getElementById("createBulkBtn")!.addEventListener("click", createBulkUsers);
    document.getElementById("getBtn")!.addEventListener("click", getUserById);
    document.getElementById("listBtn")!.addEventListener("click", listAllUsers);
    document.getElementById("listActiveBtn")!.addEventListener("click", listActiveUsers);
    document.getElementById("filterBtn")!.addEventListener("click", filterUsersByAgeName);
    document.getElementById("updateBtn")!.addEventListener("click", updateUser);
    document.getElementById("deleteBtn")!.addEventListener("click", deleteUser);
    document.getElementById("deleteAllBtn")!.addEventListener("click", deleteAllUsers);
    document.getElementById("clearBtn")!.addEventListener("click", () => {
        document.getElementById("output")!.innerHTML = "";
    });

    log("Entlite Demo Ready!");
    log("Generated TypeScript types and Connect client working!");
});
