import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { UserService } from "./gen/schema_pb.js";
import { createHash, randomFullName, toString } from "./utils.js";

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
    client.create({
        email: email,
        name: fullName,
        age: Math.ceil(Math.random() * 100),
        isAdmin: false,
        lastLoginMs: BigInt(Date.now()),
    })
    .then((response) => {
        log("✓ User created:", response);
    })
    .catch((error) => {
        log("✗ Error creating user:", toString(error));
    });
}

function getUser() {
    const idInput = document.getElementById("getId") as HTMLInputElement;
    const id = parseInt(idInput.value);
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid user ID");
        return;
    }
    log(`Getting user ${id}...`);
    client.get({ ID: id })
    .then((response) => {
        log("✓ User retrieved:", response);
    })
    .catch((error) => {
        log("✗ Error getting user:", toString(error));
    });
}

function listUsers() {
    log("Listing users...");
    client.list({ limit: 10, offset: 0 })
    .then((response) => {
        log(`✓ Users listed (${response.users.length} users):`);
        response.users.forEach((user, index) => {
            log(`  User ${index + 1}:`);
            log(`    - ID: ${user.ID}`);
            log(`    - Email: ${user.email}`);
            log(`    - Name: ${user.name}`);
            log(`    - Age: ${user.age}`);
            log(`    - Score: ${user.score}`);
            log(`    - UUID: ${user.uuid}`);
            log(`    - Is Admin: ${user.isAdmin}`);
            log(`    - Last Login: ${user.lastLoginMs}`);
        });
    })
    .catch((error) => {
        log("✗ Error listing users:", toString(error));
    });
}

function updateUser() {
    const idInput = document.getElementById("updateId") as HTMLInputElement;
    const id = parseInt(idInput.value);
    if (isNaN(id) || id <= 0) {
        log("✗ Invalid user ID");
        return;
    }
    log(`Updating user ${id}...`);
    client.update({
        ID: id,
        email: "updated@example.com",
        name: "Updated User",
        age: 30,
        isAdmin: true,
        lastLoginMs: BigInt(Date.now()),
    })
    .then((response) => {
        log("✓ User updated:");
        log(`  - ID: ${response.ID}`);
        log(`  - Email: ${response.email}`);
        log(`  - Name: ${response.name}`);
        log(`  - Age: ${response.age}`);
        log(`  - Score: ${response.score}`);
        log(`  - UUID: ${response.uuid}`);
        log(`  - Is Admin: ${response.isAdmin}`);
        log(`  - Last Login: ${response.lastLoginMs}`);
    })
    .catch((error) => {
        log("✗ Error updating user:", toString(error));
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
    client.delete({ ID: id })
    .then((response) => {
        log("✓ User deleted:", response);
    })
    .catch((error) => {
        log("✗ Error deleting user:", toString(error));
    });
}

document.addEventListener("DOMContentLoaded", () => {
    console.log("DOMContentLoaded", document.getElementById("createBtn"));
    document.getElementById("createBtn")!.addEventListener("click", createUser);
    document.getElementById("getBtn")!.addEventListener("click", getUser);
    document.getElementById("listBtn")!.addEventListener("click", listUsers);
    document.getElementById("updateBtn")!.addEventListener("click", updateUser);
    document.getElementById("deleteBtn")!.addEventListener("click", deleteUser);
    document.getElementById("clearBtn")!.addEventListener("click", () => {
        document.getElementById("output")!.innerHTML = "";
    });

    log("Entlite Demo Ready!");
    log("Generated TypeScript types and Connect client working!");
});
