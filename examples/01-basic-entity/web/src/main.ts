import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { UserService } from "./gen/schema_pb.js";

window.onload = async () => {
    const transport = createConnectTransport({
        baseUrl: "http://localhost:8080",
    });

    const client = createClient(UserService, transport);

    function log(message: string, data?: any) {
        console.log(message, data);
        const output = document.getElementById("output")!;
        const line = document.createElement("div");
        line.textContent = data 
            ? `${message} ${JSON.stringify(data, null, 2)}`
            : message;
        output.appendChild(line);
    }

    async function createUser() {
        try {
            log("Creating user...");
            const response = await client.create({
            email: "test@example.com",
            name: "Test User",
            age: 25,
            isAdmin: false,
            lastLoginMs: BigInt(Date.now()),
            });
            log("✓ User created:", response);
            log(`  - ID: ${response.id}`);
            log(`  - Email: ${response.email}`);
            log(`  - Name: ${response.name}`);
            log(`  - Score: ${response.score}`);
            log(`  - UUID: ${response.uuid}`);
            log(`  - API Key: ${response.apiKey}`);
            log(`  - Last Login: ${response.lastLoginMs}`);
        } catch (error) {
            log("✗ Error creating user:", error);
        }
    }

    async function getUser() {
        try {
            log("Getting user...");
            const response = await client.get({ id: 1 });
            log("✓ User retrieved:");
            log(`  - ID: ${response.id}`);
            log(`  - Email: ${response.email}`);
            log(`  - Name: ${response.name}`);
            log(`  - Age: ${response.age}`);
            log(`  - Score: ${response.score}`);
            log(`  - UUID: ${response.uuid}`);
            log(`  - Is Admin: ${response.isAdmin}`);
            log(`  - API Key length: ${response.apiKey.length} bytes`);
            log(`  - Last Login: ${response.lastLoginMs}`);
            log(`  - Created At: ${response.createdAt}`);
            log(`  - Updated At: ${response.updatedAt}`);
        } catch (error) {
            log("✗ Error getting user:", error);
        }
    }

    async function listUsers() {
        try {
            log("Listing users...");
            const response = await client.list({ limit: 10, offset: 0 });
            log(`✓ Users listed (${response.users.length} users):`);
            response.users.forEach((user, index) => {
                log(`  User ${index + 1}:`);
                log(`    - ID: ${user.id}`);
                log(`    - Email: ${user.email}`);
                log(`    - Name: ${user.name}`);
                log(`    - Age: ${user.age}`);
                log(`    - Score: ${user.score}`);
                log(`    - UUID: ${user.uuid}`);
                log(`    - Is Admin: ${user.isAdmin}`);
                log(`    - Last Login: ${user.lastLoginMs}`);
            });
        } catch (error) {
            log("✗ Error listing users:", error);
        }
    }

    async function updateUser() {
        try {
            log("Updating user...");
            const response = await client.update({
            id: 1,
            email: "updated@example.com",
            name: "Updated User",
            age: 30,
            isAdmin: true,
            lastLoginMs: BigInt(Date.now()),
            });
            log("✓ User updated:");
            log(`  - ID: ${response.id}`);
            log(`  - Email: ${response.email}`);
            log(`  - Name: ${response.name}`);
            log(`  - Age: ${response.age}`);
            log(`  - Score: ${response.score}`);
            log(`  - UUID: ${response.uuid}`);
            log(`  - Is Admin: ${response.isAdmin}`);
            log(`  - Last Login: ${response.lastLoginMs}`);
        } catch (error) {
            log("✗ Error updating user:", error);
        }
    }

    async function deleteUser() {
        try {
            log("Deleting user...");
            const response = await client.delete({ id: 1 });
            log("✓ User deleted:", response);
        } catch (error) {
            log("✗ Error deleting user:", error);
        }
    }

    document.addEventListener("DOMContentLoaded", () => {
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
}
