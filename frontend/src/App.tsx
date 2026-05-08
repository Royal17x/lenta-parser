import { useRef, useEffect } from 'react'
import { useParserStream } from './hooks/useParserStream'
import type { LogEntry, Product } from './types'

export default function App() {
  const { products, log, running, done, start } = useParserStream()

  return (
      <div className="min-h-screen bg-zinc-50 text-zinc-900 font-sans">
        <header className="bg-white border-b border-zinc-200 px-6 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-lg font-semibold tracking-tight">Lenta Parser</h1>
            <p className="text-xs text-zinc-400 mt-0.5">ТК124 · Москва, ТРЦ Мозаика</p>
          </div>
          <button
              onClick={start}
              disabled={running}
              className="px-4 py-2 rounded-lg text-sm font-medium bg-blue-600 text-white
                     hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed
                     transition-colors"
          >
            {running ? 'Сбор данных...' : 'Запустить'}
          </button>
        </header>

        <main className="max-w-6xl mx-auto px-6 py-6 space-y-6">
          {(log.length > 0 || running) && <LogFeed log={log} running={running} />}
          {products.length > 0 && <ProductTable products={products} />}
          {!running && !done && log.length === 0 && <EmptyState />}
        </main>
      </div>
  )
}

function EmptyState() {
  return (
      <div className="text-center py-20 text-zinc-400">
        <div className="text-4xl mb-3">🛒</div>
        <p className="text-sm">Нажмите «Запустить» чтобы собрать данные</p>
        <p className="text-xs mt-1 text-zinc-300">5 категорий · магазин ТК124</p>
      </div>
  )
}

function LogFeed({ log, running }: { log: LogEntry[]; running: boolean }) {
  const bottomRef = useRef<HTMLDivElement>(null)
  useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [log])

  return (
      <div className="bg-white border border-zinc-200 rounded-xl overflow-hidden">
        <div className="px-4 py-3 border-b border-zinc-100 flex items-center gap-2">
          {running && <span className="w-2 h-2 rounded-full bg-blue-500 animate-pulse" />}
          <span className="text-sm font-medium">Лог</span>
        </div>
        <div className="p-4 space-y-1.5 max-h-48 overflow-y-auto font-mono text-xs">
          {log.map(entry => (
              <div key={entry.id} className="flex gap-3 items-start">
                <span className="text-zinc-300 shrink-0">{entry.time}</span>
                <span className={
                  entry.type === 'error' ? 'text-red-500' :
                      entry.type === 'done'  ? 'text-green-600 font-semibold' :
                          'text-zinc-600'
                }>
              {entry.text}
            </span>
              </div>
          ))}
          <div ref={bottomRef} />
        </div>
      </div>
  )
}

function ProductTable({ products }: { products: Product[] }) {
  return (
      <div className="bg-white border border-zinc-200 rounded-xl overflow-hidden">
        <div className="px-4 py-3 border-b border-zinc-100 flex items-center justify-between">
          <span className="text-sm font-medium">Товары</span>
          <span className="text-xs text-zinc-400 bg-zinc-100 px-2 py-0.5 rounded-full">
          {products.length} шт
        </span>
        </div>
          <div className="px-4 py-2 bg-amber-50 border-b border-amber-100 text-xs text-amber-700 flex items-center gap-1.5">
              <span>ℹ️</span>
              <span>Для перехода к товару нужно открыть <a href="https://lenta.com" target="_blank" className="underline">lenta.com</a> и выбрать магазин ТК124</span>
          </div>
          <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
            <tr className="text-xs text-zinc-400 border-b border-zinc-100">
              <th className="text-left px-4 py-2.5 font-medium">Название</th>
              <th className="text-right px-4 py-2.5 font-medium">Цена</th>
              <th className="text-right px-4 py-2.5 font-medium">Обычная</th>
              <th className="text-right px-4 py-2.5 font-medium">Рейтинг</th>
              <th className="px-4 py-2.5"></th>
            </tr>
            </thead>
            <tbody>
            {products.map(p => (
                <tr key={p.id} className="border-b border-zinc-50 hover:bg-zinc-50 transition-colors">
                  <td className="px-4 py-2.5 max-w-xs">
                    <div className="truncate font-medium text-zinc-800">{p.display.name}</div>
                    <div className="text-xs text-zinc-400">{p.display.package}</div>
                  </td>
                  <td className="px-4 py-2.5 text-right font-semibold text-zinc-900 whitespace-nowrap">
                    {(p.prices.price / 100).toFixed(2)} ₽
                    {p.prices.isPromoactionPrice && (
                        <span className="ml-1.5 text-xs bg-red-100 text-red-600 px-1.5 py-0.5 rounded">акция</span>
                    )}
                  </td>
                  <td className="px-4 py-2.5 text-right text-zinc-400 text-xs whitespace-nowrap">
                    {(p.prices.priceRegular / 100).toFixed(2)} ₽
                  </td>
                  <td className="px-4 py-2.5 text-right text-xs text-zinc-500 whitespace-nowrap">
                    ★ {p.rating.rate} <span className="text-zinc-300">({p.rating.votes})</span>
                  </td>
                  <td className="px-4 py-2.5">
                    <a
                        href={`https://lenta.com/product/${p.slug}/?storeId=${p.storeId}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        title="Откройте lenta.com и выберите магазин ТК124 для корректного отображения"
                        className="text-blue-500 hover:text-blue-700 text-xs"
                    >
                      →
                    </a>
                  </td>
                </tr>
            ))}
            </tbody>
          </table>
        </div>
      </div>
  )
}