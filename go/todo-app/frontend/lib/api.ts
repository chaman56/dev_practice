const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

export interface Todo {
  id: number
  title: string
  completed: boolean
  createdAt: string
  updatedAt: string
}

async function request(path: string, opts: RequestInit = {}): Promise<Response> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...opts.headers,
    },
    ...opts,
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`)
  }
  return res
}

export async function fetchTodos(): Promise<Todo[]> {
  const res = await request("/api/todos")
  return res.json()
}

export async function createTodo(title: string): Promise<Todo> {
  const res = await request("/api/todos", {
    method: "POST",
    body: JSON.stringify({ title }),
  })
  return res.json()
}

export async function updateTodo(
  id: number,
  body: { title?: string; completed?: boolean },
): Promise<Todo> {
  const res = await request(`/api/todos/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  })
  return res.json()
}

export async function deleteTodo(id: number): Promise<void> {
  await request(`/api/todos/${id}`, { method: "DELETE" })
}
