# VeloRoute Design System

## Pattern: Real-Time Monitoring Dashboard
- **Style**: Data-Dense Dashboard / DevOps Operations
- **Best for**: Reverse proxy monitoring, load balancer observability
- **Layout**: Header + metrics cards + charts + backend table + live log feed

## Colors
| Token | Value | Usage |
|-------|-------|-------|
| Background | `#0f1117` | Page background |
| Card | `#1a1d27` | Card/panel background |
| Border | `#2a2d3e` | Borders, dividers |
| Primary | `#6366f1` | Accent, links, primary actions |
| Success | `#22c55e` | Alive status, 2xx badges |
| Error | `#ef4444` | Dead status, 5xx badges |
| Warning | `#f59e0b` | 4xx badges, high latency |
| Text Primary | `#f1f5f9` | Headings, values |
| Text Muted | `#64748b` | Labels, secondary text |

## Typography
- **UI**: Inter, system-ui, sans-serif
- **Logs**: JetBrains Mono, monospace

## Effects
- Smooth transitions 150-300ms
- Subtle hover states on interactive rows
- Pulse animation on alive status dots
- Fade-in on new log entries
- `prefers-reduced-motion`: disable animations

## Components
- Skeleton loaders on initial load (not spinners)
- Lucide React icons only
- Responsive: 2-col mobile, 4-col desktop for metrics cards

## Anti-patterns
- No emoji as icons
- No inline styles — Tailwind only
- No spinners for data loading
