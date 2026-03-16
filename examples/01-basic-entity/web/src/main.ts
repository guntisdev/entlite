import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { UserService } from "./gen/schema_connect.js";

// Create transport
const transport = createConnectTransport({
  baseUrl: "http://localhost:8080",
});

// Create client
const client = createClient(UserService, transport);

// Helper to log to both console and page
function log(message: string, data?: any) {
  console.log(message, data);
  const output = document.getElementById("output")!;
  const line = document.createElement("div");
  line.textContent = data 
    ? `${message} ${JSON.stringify(data, null, 2)}`
    : message;
  output.appendChild(line);
}

// Demo functions
async function createUser() {
  try {
    log("Creating user...");
    const response = await client.create({
      email: "test@example.com",
      name: "Test User",
      age: 25,
      isAdmin: false,
    });
    log("✓ User created:", response);
  } catch (error) {
    log("✗ Error creating user:", error);
  }
}

async function getUser() {
  try {
    log("Getting user...");
    const response = await client.get({ id: BigInt(1) });
    log("✓ User retrieved:", response);
  } catch (error) {
    log("✗ Error getting user:", error);
  }
}

async function listUsers() {
  try {
    log("Listing users...");
    const response = await client.list({ limit: 10, offset: 0 });
    log("✓ Users listed:", response);
  } catch (error) {
    log("✗ Error listing users:", error);
  }
}

async function updateUser() {
  try {
    log("Updating user...");
    const response = await client.update({
      id: BigInt(1),
      email: "updated@example.com",
      name: "Updated User",
      age: 30,
      isAdmin: true,
    });
    log("✓ User updated:", response);
  } catch (error) {
    log("✗ Error updating user:", error);
  }
}

async function deleteUser() {
  try {
    log("Deleting user...");
    const response = await client.delete({ id: BigInt(1) });
    log("✓ User deleted:", response);
  } catch (error) {
    log("✗ Error deleting user:", error);
  }
}

// Set up UI
document.addEventListener("DOMContentLoaded", () => {
  document.getElementById("createBtn")!.addEventListener("click", createUser);
  document.getElementById("getBtn")!.addEventListener("click", getUser);
  document.getElementById("listBtn")!.addEventListener("click", listUsers);
  document.getElementById("updateBtn")!.addEventListener("click", updateUser);
  document.getElementById("deleteBtn")!.addEventListener("click", deleteUser);
  document.getElementById("clearBtn")!.addEventListener("click", () => {
    document.getElementById("output")!.innerHTML = "";
  });

  log("🚀 Entlite Demo Ready!");
  log("Generated TypeScript types and Connect client working!");
});
