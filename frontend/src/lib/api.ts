// Single typed client for the backend's /api/* routes — mirrors the Item
// JSON shape in internal/triage/item.go. Update both together.
export interface Item {
  id: string
  artistGuess: string
  filename: string
  sizeBytes: number
  modTime: string
  rawTitle: string
  description: string
  uploader: string
  category: string
  inferredArtist: string
  inferredArtistConfidence: number
  inferredTitle: string
  inferredTitleConfidence: number
  inferredIsFullSet: boolean
  inferredSource: string
  hasThumbnail: boolean
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, init)
  if (!res.ok) {
    const text = await res.text().catch(() => "")
    throw new Error(text || `${res.status} ${res.statusText}`)
  }
  if (res.status === 204) {
    return undefined as T
  }
  return res.json() as Promise<T>
}

export const api = {
  listItems: () => request<Item[]>("/api/items"),

  thumbnailUrl: (id: string) => `/api/items/${id}/thumbnail`,
  streamUrl: (id: string) => `/api/items/${id}/stream`,

  accept: (id: string, artist: string, title: string) =>
    request<{ path: string }>(`/api/items/${id}/accept`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ artist, title }),
    }),

  reject: (id: string) =>
    request<void>(`/api/items/${id}`, { method: "DELETE" }),
}
