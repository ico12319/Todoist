Todoist
The classic todo list app should support creating different lists of todos. Each list should have its own name and description. In the lists, there are todos that are separate from the others. The access to lists should be managed in the database by user emails. Users should be able to log in to the application via a GitHub account. The users should be part of a specific GitHub organization in order to be able to log in.

Implementation

Implement a Todo microservice that supports the following:

Login with GitHub
Create/Modify/Delete todo lists and todos
Save Todo lists and todos in the database
Requirements for Todo Lists

One user can participate in many list and many lists can be shared with one user(many to many realationship).
One user can be assigned to multiple todos and one todo can only one assignee(one to many realtionship).
One list can have many todos and one todo can be part of only one list(one to many realtionship).
One user can be owner of many lists and one list can have only one owner(one to many relationship).

Only Admins and Writers can create new entities(todos and lists).
Each list has an owner (the user who created the list).
Owners, Admins and collaborators of the list can modify todos in the list.
Owners, Admins and collaborators can add new users to the todo list.
Only users that are part of a certain Todo list or are Admins can modify it.

Everyone who is authorized with their JWT can read lists, todos and users and infomration related to these entities.
Requirements for GitHub Organization

The GitHub organization should have 3 teams - Todod-app-Readers, Todo-app-Writers, and Todo-app-Admins and depending on the one you are participating in your role in the API will be determined
and encoded into your JWT.
The flow is straighforward. You make a GET Http request at /github/login then you are asked to log in into your github account after you enter your credentials and give access to your private info
you will be redirected to /auth2/callback where you will be issued a JWT and a Refresh token. After that you just grab you JWT and put into the AUTH header in Postman and you are
authorized to use the API. JWT are available for only 3 minutes then they expire. When you token expire you should make a POST Http request to /tokens/refresh where you should put 
your Refresh token in the request body and if it's correct you will be given a new JWT and a new Refresh token, this is done to add even more security.
I have two implementations. REST API and GraphQL API. The GraphQL API is actually a facade that just calls the REST API under the hood. The GraphQL Schema is written by me and 
have implemented interfaces, enums, directives, queries and mutaions. Both REST and GraphQL support cursor based pagination(limit:x, after: id).

REST API endpoints:
GET /lists - read all lists.
GET /todos - read all todos.
GET /users - read all users.
GET /lists/{list_id} - receive information about specific list.
GET /todos/{todos_id} - receive information about specific todo.
GET user/{user_id} - receive information about specific user.
GET /lists/{list_id}/collaborators - see all users that are collaborators in the current list.
GET /lists/{list_id}/todos - see all todos that are part of the current list.
GET /lists/{list_id}/owner - see the owner of the current list.
GET /todos/{todo_id}/assignee - see the user assigned to the current todo.
GET /users/{user_id}/lists - see all the lists where the current user is collaborator.
GET /users/{user_id}/todos - see all todos that are assigned to the current user.
POST /lists - create list.
POST /todos - create todo.
DELETE /lists - delete all lists(admins  only).
DELETE /todos - delete all todos(admins only).
DELETE /users - delete all users(admins only).
DELETE /lists/{list_id} - delete current list.
DELETE /todos/{todo_id} - delete current todo.
DELETE /users/{user_id} - delete current user.
PATCH /lists/{list_id} - update current list.
PATH /todos/{todo_id} - update current todo.
POST /token/refresh -  get new JWT and Refresh tokens.
GET /github/login - login with GitHub.
