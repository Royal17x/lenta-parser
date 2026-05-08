import { useState, useCallback, useRef } from 'react'
import type { Product, LogEntry, SSEEvent } from '../types'

export function useParserStream() {
    const [products, setProducts] = useState<Product[]>([])
    const [log, setLog] = useState<LogEntry[]>([])
    const [running, setRunning] = useState(false)
    const [done, setDone] = useState(false)
    const logIdRef = useRef(0)

    const addLog = useCallback((type: LogEntry['type'], text: string) => {
        const entry: LogEntry = {
            id: logIdRef.current++,
            type,
            text,
            time: new Date().toLocaleTimeString('ru-RU'),
        }
        setLog(prev => [...prev, entry])
    }, [])

    const start = useCallback(async () => {
        setRunning(true)
        setDone(false)
        setProducts([])
        setLog([])

        await fetch('/api/start', { method: 'POST' })

        const es = new EventSource('/api/stream')

        es.onmessage = (e) => {
            const event: SSEEvent = JSON.parse(e.data)

            if (event.type === 'progress') {
                addLog('progress', `${event.category}: собрано ${event.count} товаров`)
            } else if (event.type === 'error') {
                addLog('error', `Ошибка: ${event.message}`)
            } else if (event.type === 'done') {
                addLog('done', `Готово — ${event.count} товаров`)
                setRunning(false)
                setDone(true)
                es.close()
                fetch('/api/results')
                    .then(r => r.json())
                    .then(setProducts)
            }
        }

        es.onerror = () => {
            addLog('error', 'Соединение прервано')
            setRunning(false)
            es.close()
        }
    }, [addLog])

    return { products, log, running, done, start }
}