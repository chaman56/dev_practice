"use client"

import { updateTodo, deleteTodo } from "@/lib/api"
import type { Todo } from "@/lib/api"
import { useState } from "react"

interface Props {
  todo: Todo
  onChange: (todo: Todo) => void
  onDelete: (id: number) => void
}

export default function TodoItem({ todo, onChange, onDelete }: Props) {
  const [editing, setEditing] = useState(false)
  const [editTitle, setEditTitle] = useState(todo.title)

  const toggle = async () => {
    const updated = await updateTodo(todo.id, { completed: !todo.completed })
    onChange(updated)
  }

  const save = async () => {
    const trimmed = editTitle.trim()
    if (trimmed === "") return
    const updated = await updateTodo(todo.id, { title: trimmed })
    onChange(updated)
    setEditing(false)
  }

  const handleRemove = async () => {
    await deleteTodo(todo.id)
    onDelete(todo.id)
  }

  return (
    <div className="flex items-center gap-3 p-3 bg-white rounded-lg shadow-sm border border-gray-100">
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={toggle}
        className="w-5 h-5 text-blue-600 rounded focus:ring-blue-500 cursor-pointer"
        aria-label="Mark complete"
      />
      {editing ? (
        <input
          autoFocus
          value={editTitle}
          onChange={(e) => setEditTitle(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") save()
          }}
          onBlur={save}
          className="flex-1 px-2 py-1 border border-blue-500 rounded text-lg focus:outline-none"
        />
      ) : (
        <div
          className="flex-1 cursor-pointer select-none"
          onDoubleClick={() => setEditing(true)}
          title="Double-click to edit"
        >
          <span
            className={`block ${todo.completed ? "line-through text-gray-400" : "text-gray-800"}`}
          >
            {todo.title}
          </span>
        </div>
      )}
      <button
        onClick={handleRemove}
        className="px-3 py-1 text-xs bg-red-100 text-red-700 rounded hover:bg-red-200 transition"
        aria-label="Delete todo"
      >
        Delete
      </button>
    </div>
  )
}
