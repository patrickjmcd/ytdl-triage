import { useEffect, useState } from "react"
import { Loader2, PartyPopper, RefreshCw } from "lucide-react"
import { toast } from "sonner"
import { api, type Item } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ItemCard } from "@/components/item-card"

function App() {
  const [items, setItems] = useState<Item[] | null>(null)
  const [refreshing, setRefreshing] = useState(false)

  async function load() {
    setRefreshing(true)
    try {
      setItems(await api.listItems())
    } catch (e) {
      toast.error(`Failed to load items: ${(e as Error).message}`)
    } finally {
      setRefreshing(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function remove(id: string) {
    setItems((prev) => prev?.filter((i) => i.id !== id) ?? prev)
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <header className="mb-6 flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">ytdl-triage</h1>
          <p className="text-sm text-muted-foreground">
            Downloads the postprocessor wasn't confident enough to file on
            its own.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {items && <Badge variant="outline">{items.length} pending</Badge>}
          <Button variant="outline" size="icon" onClick={load} disabled={refreshing}>
            <RefreshCw className={refreshing ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
          </Button>
        </div>
      </header>

      {items === null ? (
        <div className="flex justify-center py-24 text-muted-foreground">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : items.length === 0 ? (
        <div className="flex flex-col items-center gap-3 py-24 text-muted-foreground">
          <PartyPopper className="h-10 w-10" />
          <p>Nothing needs review.</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((item) => (
            <ItemCard
              key={item.id}
              item={item}
              onAccepted={remove}
              onRejected={remove}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export default App
