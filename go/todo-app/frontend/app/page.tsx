"use client"

import { useEffect, useState } from "react"
import { fetchTodos } from "@/lib/api"
import type { Todo } from "@/lib/api"
import CreateTodo from "@/components/CreateTodo"
import TodoItem from "@/components/TodoItem"

export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchTodos()
      .then(setTodos)
      .catch((err) => {
        console.error(err)
      })
      .finally(() => setLoading(false))
  }, [])

  const handleAdd = (todo: Todo) => {
    setTodos((prev) => [todo, ...prev])
  }

  const handleUpdate = (updated: Todo) => {
    setTodos((prev) => prev.map((t) => (t.id === updated.id ? updated : t)))
  }

  const handleDelete = (id: number) => {
    setTodos((prev) => prev.filter((t) => t.id !== id))
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto p-6 flex items-center justify-center min-h-[40vh]">
        Loading todos…
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="flex items-center gap-2 mb-6">
        <h1 className="text-2xl font-bold">Todo List</h1>
      </div>
      <CreateTodo onAdd={handleAdd} />
      <div className="space-y-2">
        {todos.length === 0 ? (
          <p className="mt-4 text-gray-500">No todos yet. Add one above!</p>
        ) : (
          todos.map((todo) => (
            <TodoItem
              key={todo.id}
              todo={todo}
              onChange={handleUpdate}
              onDelete={handleDelete}
            />
          ))
        )}
      </div>
    </div>
  )
}
