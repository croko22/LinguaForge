import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { fetchStats, type ReviewActivity } from '../api/stats'
import { fetchDocuments } from '../api/documents'
import { useDueCount } from '../hooks/useReview'

function RecentBooks() {
  const { data: docs } = useQuery({
    queryKey: ['documents'],
    queryFn: fetchDocuments,
  })

  const recent = docs?.slice(0, 5) ?? []

  if (recent.length === 0) return null

  return (
    <div className="bg-surface rounded-xl border border-border p-5">
      <h2 className="text-sm font-semibold text-text mb-4">Recent Books</h2>
      <div className="space-y-3">
        {recent.map((doc) => (
          <Link
            key={doc.id}
            to={`/read/${doc.id}/0`}
            className="flex items-center justify-between p-3 rounded-lg hover:bg-surface-hover transition-colors group"
          >
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-text truncate group-hover:text-primary transition-colors">
                {doc.title}
              </p>
              <p className="text-xs text-text-muted mt-0.5">
                {doc.language} · {doc.chapter_count} chapters
              </p>
            </div>
            <span className="text-xs text-text-muted shrink-0 ml-3">
              {new Date(doc.created_at).toLocaleDateString()}
            </span>
          </Link>
        ))}
      </div>
      <Link
        to="/library"
        className="mt-3 block text-center text-xs text-primary hover:underline font-medium"
      >
        View all books →
      </Link>
    </div>
  )
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="bg-surface rounded-xl border border-border p-5">
      <p className={`text-3xl font-bold ${color}`}>{value}</p>
      <p className="text-xs text-text-muted mt-1">{label}</p>
    </div>
  )
}

function LanguageBreakdown({ counts }: { counts: { language: string; count: number }[] }) {
  const total = counts.reduce((s, c) => s + c.count, 0)
  if (counts.length === 0) return null

  return (
    <div className="bg-surface rounded-xl border border-border p-5">
      <h2 className="text-sm font-semibold text-text mb-3">Languages</h2>
      <div className="space-y-2">
        {counts.map((c) => (
          <div key={c.language} className="flex items-center gap-3">
            <span className="text-sm text-text w-20 truncate capitalize">{c.language}</span>
            <div className="flex-1 h-2 bg-border rounded-full overflow-hidden">
              <div
                className="h-full bg-primary rounded-full transition-all"
                style={{ width: `${(c.count / total) * 100}%` }}
              />
            </div>
            <span className="text-xs text-text-muted tabular-nums w-8 text-right">{c.count}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

function ReviewHeatmap({ activity }: { activity: ReviewActivity[] }) {
  const activityMap = new Map(activity.map((a) => [a.date, a.count]))
  const maxCount = Math.max(...activity.map((a) => a.count), 1)

  const today = new Date()
  const cells: { date: string; count: number; level: number }[] = []
  for (let i = 363; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(d.getDate() - i)
    const dateStr = d.toISOString().slice(0, 10)
    const count = activityMap.get(dateStr) ?? 0
    const level = count === 0 ? 0 : Math.min(Math.ceil((count / maxCount) * 4), 4)
    cells.push({ date: dateStr, count, level })
  }

  const weeks: typeof cells[] = []
  for (let i = 0; i < cells.length; i += 7) {
    weeks.push(cells.slice(i, i + 7))
  }

  const levelColors = [
    'bg-border',
    'bg-primary-light',
    'bg-primary/40',
    'bg-primary/70',
    'bg-primary',
  ]

  const monthLabels: { index: number; label: string }[] = []
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  weeks.forEach((week, wi) => {
    const d = new Date(week[0].date)
    const month = d.getMonth()
    if (wi === 0 || new Date(weeks[wi - 1][0].date).getMonth() !== month) {
      monthLabels.push({ index: wi, label: months[month] })
    }
  })

  return (
    <div className="bg-surface rounded-xl border border-border p-5">
      <h2 className="text-sm font-semibold text-text mb-3">Review Activity</h2>
      {cells.length === 0 || activity.length === 0 ? (
        <p className="text-xs text-text-muted">No review activity yet. Start reading and reviewing words!</p>
      ) : (
        <div className="overflow-x-auto">
          <div className="flex gap-0.5" style={{ minWidth: weeks.length * 14 }}>
            <div className="flex flex-col gap-0.5 pr-1 pt-5">
              {['Mon', '', 'Wed', '', 'Fri', '', 'Sun'].map((d) => (
                <span key={d} className="text-[10px] text-text-muted leading-3 h-3">{d}</span>
              ))}
            </div>
            <div>
              <div className="flex gap-0.5 mb-1">
                {monthLabels.map((m) => (
                  <span
                    key={m.index}
                    className="text-[10px] text-text-muted"
                    style={{ marginLeft: m.index === 0 ? 0 : (m.index - (monthLabels[monthLabels.indexOf(m) - 1]?.index ?? 0)) * 14 - 14 }}
                  >
                    {m.label}
                  </span>
                ))}
              </div>
              <div className="flex gap-0.5">
                {weeks.map((week, wi) => (
                  <div key={wi} className="flex flex-col gap-0.5">
                    {week.map((cell) => (
                      <div
                        key={cell.date}
                        className={`w-3 h-3 rounded-sm ${levelColors[cell.level]} transition-colors`}
                        title={`${cell.date}: ${cell.count} reviews`}
                      />
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-1 mt-2 justify-end text-[10px] text-text-muted">
            <span>Less</span>
            {levelColors.map((c, i) => (
              <div key={i} className={`w-3 h-3 rounded-sm ${c}`} />
            ))}
            <span>More</span>
          </div>
        </div>
      )}
    </div>
  )
}

export default function DashboardPage() {
  const { data: stats, isLoading, isError } = useQuery({
    queryKey: ['stats'],
    queryFn: fetchStats,
  })
  const { data: dueCount } = useDueCount()

  if (isLoading) {
    return (
      <div className="max-w-5xl mx-auto px-4 py-8">
        <p className="text-text-secondary animate-pulse">Loading dashboard...</p>
      </div>
    )
  }

  if (isError || !stats) {
    return (
      <div className="max-w-5xl mx-auto px-4 py-8">
        <p className="text-danger">Failed to load dashboard stats</p>
      </div>
    )
  }

  return (
    <div className="max-w-5xl mx-auto px-4 py-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-text">Dashboard</h1>
        {dueCount !== undefined && dueCount > 0 && (
          <Link
            to="/review"
            className="flex items-center gap-2 px-4 py-2 bg-primary text-text-inverse rounded-lg hover:bg-primary-hover transition-colors text-sm font-medium"
          >
            Review due ({dueCount})
          </Link>
        )}
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard label="Books" value={stats.total_documents} color="text-primary" />
        <StatCard label="Words" value={stats.total_words} color="text-grade-good" />
        <StatCard label="Chapters" value={stats.total_chapters} color="text-grade-hard" />
        <StatCard
          label="Due for review"
          value={dueCount ?? 0}
          color={dueCount && dueCount > 0 ? 'text-danger' : 'text-text-muted'}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <RecentBooks />
        <LanguageBreakdown counts={stats.language_counts} />
      </div>

      <ReviewHeatmap activity={stats.review_activity} />
    </div>
  )
}
