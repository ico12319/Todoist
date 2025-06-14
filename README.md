# Todoist Microservice

A production-ready microservice that powers **Todoist-style** lists and todos with fine-grained role-based access, OAuth2 login via GitHub, and both **REST** & **GraphQL** APIs.

---

## ✨ Key Features

* **GitHub OAuth2** – Only members of a specific GitHub organisation can sign in.  
  Role is inferred from their team (`Todo-app-Readers`, `Todo-app-Writers`, `Todo-app-Admins`).
* **JWT & Refresh Tokens** – Short-lived (3 min) access tokens, long-lived refresh tokens, automatic rotation.
* **Dual API Surface** – Hand-crafted GraphQL façade atop the REST API.  
  Both expose **cursor-based pagination** (`limit`, `after`).
* **Role-Aware Authorisation** – Declarative middleware enforces all business rules (owner, collaborator, writer, admin…).
* **Clean Architecture** – Each domain entity (`Todo`, `List`, `User`) has its own **Repository → Service → Handler** pipeline.
* **Pluggable Decorators** – SQL decorators (pagination, filtering, sorting) & URL decorators built via factories.
* **Middleware-First** – Cross-cutting concerns (auth, logging, tracing, validation) extracted into reusable middleware.
* **REST Endpoints**

* **GET /lists – Read all lists (Reader)**

* **GET /todos – Read all todos (Reader)**

* **GET /users – Read all users (Reader)**

* **GET /lists/{list_id} – Info about a specific list (Reader)**

* **GET /todos/{todo_id} – Info about a specific todo (Reader)**

* **GET /users/{user_id} – Info about a specific user (Reader)**

* **GET /lists/{list_id}/collaborators – Collaborators in a list (Reader)**

* **GET /lists/{list_id}/todos – Todos in a list (Reader)**

* **GET /lists/{list_id}/owner – Owner of a list (Reader)**

* **GET /todos/{todo_id}/assignee – Assignee of a todo (Reader)**

* **GET /users/{user_id}/lists – Lists where the user collaborates (Reader)**

GET /users/{user_id}/todos – Todos assigned to the user (Reader / Admin)

* **POST /lists – Create a list (Writer)**

* **POST /todos – Create a todo (Collaborator, Owner of the list where todo belongs or Admin)**

* **DELETE /lists – Delete all lists (Admin)**

* **DELETE /todos – Delete all todos (Admin)**

* **DELETE /users – Delete all users (Admin)**

* **DELETE /lists/{list_id} – Delete a list (Owner,Collaborator or Admin)**

** **DELETE /todos/{todo_id} – Delete a todo (Owner of the list where todo belongs/ Admin / Collaborator)**

** **DELETE /users/{user_id} – Delete a user (Admin or you can delete your account)**

**PATCH /lists/{list_id} – Update a list (Owner / Collaborator / Admin)**

**PATCH /todos/{todo_id} – Update a todo (Owner / Collaborator / Admin)**

**POST /token/refresh – Obtain a new JWT + refresh token (Any authenticated)**

**GET /github/login – OAuth2 login entry‑point (Public)**

GraphQL operations mirror these REST routes (see schema.graphql) but there an additional directive that only admins can see the role of the users. If you are not an admin you will
see null everytime you query this field.

---

## 🏗️ Project Structure

```text
📁 cmd/                  # Entrypoints (REST & GraphQL binaries)
│   ├── main.go
📁 internal/
    ├── graph/           # GraphQL resolvers implementation
│   ├── todos/           # Domain: Todo
│   │   ├── repository/  # DB adapters (Postgres)
│   │   ├── service/     # Business rules
│   │   └── handler/     # HTTP & GraphQL
│   ├── lists/           # Domain: List
│   ├── users/           # Domain: User
│   ├── auth/            # JWT, GitHub OAuth, middleware
│   ├── middlewares/      # Logging, recovery, rate-limit…
│   ├── decorators/      # SQL & URL decorators
│   └── config/          # Typed config loader
      
│
📁 migrations/           # SQL migration files
