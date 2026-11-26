"use client"

import { createTodo } from "@/lib/api"
import type { Todo } from "@/lib/api"
import { useState } from "react"

interface Props {
  onAdd: (todo: Todo) => void
}

export default function CreateTodo({ onAdd }: Props) {
  const [title, setTitle] = useState("")
  const [error, setError] = useState("")

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")

    try {
      const todo = await createTodo(title.trim())
      onAdd(todo)
      setTitle("")
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create todo")
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 mb-6">
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="What needs to be done?"
        className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none"
        autoFocus
      />
      <button
        type="submit"
        disabled={title.trim() === ""}
        className="px-4 py-2 bg-blue-600 text-white rounded-lg disabled:bg-gray-400 hover:!bg-blue-700 disabled:!cursor-not-allowed transition"
      >
        Add
      </button>
      {error && <span className="text-red-500 text-sm mt-2 block">{error}</span>}
    </form>
  )
}
