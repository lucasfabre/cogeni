import { ApiClient } from "./api";

async function test() {
  const client = new ApiClient("http://localhost:8000");

  console.log("Testing Todo creation...");
  const newTodo = { id: 1, title: "Learn cogeni", completed: false };
  const createdTodo = await client.createTodo(newTodo);
  console.log("Created:", createdTodo);

  if (createdTodo.title !== newTodo.title) {
    throw new Error("Todo title mismatch");
  }

  console.log("Testing User creation...");
  const newUser = { id: 1, username: "jacques", email: "jacques@example.com" };
  const createdUser = await client.createUser(newUser);
  console.log("Created:", createdUser);

  console.log("Testing list Todos...");
  const todos = await client.getTodos();
  console.log("Todos count:", todos.length);
  if (todos.length === 0) {
    throw new Error("List should not be empty");
  }

  console.log("SDK Test passed!");
}

test().catch((err) => {
  console.error("SDK Test failed:", err);
  process.exit(1);
});
