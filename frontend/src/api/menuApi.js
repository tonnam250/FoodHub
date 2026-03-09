const BASE_URL = "http://localhost:3002"

async function parseResponse(res) {
  const text = await res.text()
  let body = text
  try {
    body = text ? JSON.parse(text) : null
  } catch {
    body = text
  }

  if (!res.ok) {
    const message = typeof body === "string" ? body : body?.error || "Request failed"
    throw new Error(message)
  }

  return body
}

export async function getMenus() {
  const res = await fetch(`${BASE_URL}/menu`)
  return parseResponse(res)
}

export async function createMenuItem(payload) {
  const res = await fetch(`${BASE_URL}/menu`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
  return parseResponse(res)
}

export async function updateMenuItem(id, payload) {
  const res = await fetch(`${BASE_URL}/menu/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
  return parseResponse(res)
}

export async function deleteMenuItem(id) {
  const res = await fetch(`${BASE_URL}/menu/${id}`, {
    method: "DELETE",
  })
  return parseResponse(res)
}
