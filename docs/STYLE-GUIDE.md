# BentoTask Visual Identity — Style Guide

**Status**: ACTIVE  
**Date**: 2026-04-09  
**Primary Theme**: Bento Alt (dark mode default)  
**Archived Theme**: Clay Alt (light mode / future customization)

---

## 1. Chosen Direction: "Bento Alt"

**Aesthetic**: Cool, moody dark interface with periwinkle/violet accent. Feels like a polished developer tool meets modern productivity app — Arc browser's restraint with Things 3's clarity.

**Why it works**: The periwinkle accent (#9b7ede) is distinctive without being aggressive. The dark backgrounds have a subtle blue undertone (#0f0f12, #1a1a22) that feels warmer than pure gray. The emerald green (#09814a) for success/completion provides strong contrast against the violet primary.

---

## 2. Color System

### Dark Mode (Bento Alt — Primary)

```css
:root {
  /* Backgrounds */
  --bg-base: #0f0f12;          /* Page background — near-black with blue undertone */
  --bg-surface: #1a1a22;       /* Cards, inputs, elevated surfaces */
  --bg-elevated: #22222e;      /* Badges, tracks, nested surfaces */

  /* Borders */
  --border-default: #2e2d4d;   /* Card borders, dividers — subtle violet tint */
  --border-subtle: #3a3a4c;    /* Lighter borders for secondary elements */

  /* Text */
  --text-primary: #e5e5e7;     /* Main text — off-white, easy on eyes */
  --text-secondary: #888;      /* Metadata, labels */
  --text-tertiary: #555;       /* Placeholders, disabled text */
  --text-muted: #333;          /* Separators, very subtle text */

  /* Accent — Periwinkle / Violet */
  --accent-primary: #9b7ede;   /* Primary actions, links, active states */
  --accent-hover: #b9a4e8;     /* Hover states, highlighted text */
  --accent-subtle: rgba(155, 126, 222, 0.12);  /* Badge backgrounds, tints */

  /* Success — Emerald Green */
  --success: #09814a;          /* Completed, logged, positive actions */
  --success-subtle: rgba(9, 129, 74, 0.10);    /* Success badge bg */
  --success-border: rgba(9, 129, 74, 0.25);    /* Success button border */

  /* Warning / Danger — Deep Rose */
  --warning: #b6174b;          /* At-risk, overdue, urgent */
  --warning-subtle: rgba(182, 23, 75, 0.12);   /* Warning badge bg */

  /* Priority Colors */
  --priority-urgent-bg: rgba(182, 23, 75, 0.12);
  --priority-urgent-text: #e04878;
  --priority-high-bg: rgba(155, 126, 222, 0.12);
  --priority-high-text: #9b7ede;
  --priority-medium-bg: rgba(155, 126, 222, 0.08);
  --priority-medium-text: #8a7abc;
  --priority-low-bg: rgba(9, 129, 74, 0.08);
  --priority-low-text: #09814a;

  /* Score Visualization */
  --score-track: #22222e;
  --score-fill: #9b7ede;       /* Primary score bar */
  --score-fill-alt: #09814a;   /* Factor breakdown bars */

  /* Shadows */
  --shadow-card: 0 1px 3px rgba(0, 0, 0, 0.2);
  --shadow-elevated: 0 2px 6px rgba(0, 0, 0, 0.3);
}
```

### Light Mode (Clay Alt — Archived for Future Use)

```css
:root[data-theme="light"] {
  /* Backgrounds */
  --bg-base: #fdecef;          /* Page background — warm blush pink */
  --bg-surface: #fff5f6;       /* Cards, inputs */
  --bg-elevated: #f3e4e8;      /* Badges, tracks */

  /* Borders */
  --border-default: #e8d0d6;   /* Card borders — warm rose */
  --border-subtle: #ede0e4;

  /* Text */
  --text-primary: #0f110c;     /* Near-black */
  --text-secondary: #84828f;   /* Muted text */
  --text-tertiary: #b0a8b0;

  /* Accent — Rosewood */
  --accent-primary: #9d6381;
  --accent-hover: #7a4e66;
  --accent-subtle: rgba(157, 99, 129, 0.12);

  /* Success — Deep Berry */
  --success: #612940;
  --success-subtle: rgba(97, 41, 64, 0.10);

  /* Warning */
  --warning: #9d6381;          /* Softer in light mode */

  /* Score */
  --score-track: #e8d0d6;
  --score-fill: #9d6381;
  --score-fill-alt: #612940;

  /* Shadows */
  --shadow-card: 0 2px 8px rgba(157, 99, 129, 0.08);
}
```

---

## 3. Typography

### Font Stack
```css
/* Primary — System sans-serif, optimized per platform */
font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', Roboto, sans-serif;
```

No external font dependencies for the shipped theme. System fonts render fastest and feel native on each platform. The Clay Alt archive uses Nunito (Google Fonts) — load it only if that theme is active.

### Scale
| Element | Size | Weight | Letter-spacing |
|---------|------|--------|---------------|
| Page title (h1) | 1.5rem | 600 | -0.01em |
| Card title | 0.93–0.95rem | 500 | normal |
| Body text | 0.9rem | 400 | normal |
| Badge text | 0.6rem | 500 | normal |
| Label / caption | 0.65rem | 500 | 0.04–0.08em (uppercase) |
| Score value | 0.7rem | 600 | tabular-nums |
| Stat text | 0.7rem | 400 | normal |

### Rules
- Titles: `font-weight: 600`, slightly negative letter-spacing
- Labels/captions: uppercase, small letter-spacing, reduced opacity
- Numbers (scores, stats, durations): use `font-variant-numeric: tabular-nums` for alignment
- Never bold body text — use color/opacity for hierarchy instead

---

## 4. Spacing & Layout

### Radius Scale
| Element | Radius |
|---------|--------|
| Cards, inputs | 10px |
| Buttons | 8px |
| Badges | 5px |
| Score bars | 2–3px |
| Status indicators | 50% (circle) |

### Spacing
| Context | Value |
|---------|-------|
| Card padding | 0.8–1rem |
| Card gap (between cards) | 0.4rem |
| Section margin-bottom | 1.25rem |
| Badge padding | 0.15rem 0.45rem |
| Content max-width | 700px (views), 900px (layout) |

### Cards
- All list items are cards: `background: var(--bg-surface)`, `border: 1px solid var(--border-default)`, `border-radius: 10px`
- Subtle shadow: `var(--shadow-card)`
- No divider borders between items — card gaps provide separation

---

## 5. Component Patterns

### Checkboxes
- Circle outline: `border: 2px solid var(--accent-primary)`, `border-radius: 50%`, 18×18px
- Hover: border color brightens to `var(--accent-hover)`
- No fill on unchecked — clean open circle

### Badges
- Small pill: `background: var(--bg-elevated)`, `color: var(--text-secondary)`
- Priority badges use semantic color pairs (bg + text)
- Tags use default badge style with `#` prefix

### Score Bars
- Track: `height: 6px`, `background: var(--score-track)`, `border-radius: 3px`
- Fill: `background: var(--score-fill)`, smooth `width` transition (0.4s)
- Factor breakdown bars: thinner (4px), use `var(--score-fill-alt)` for contrast
- Value label: right-aligned, `var(--accent-primary)` color, `font-weight: 600`

### Buttons
- Primary action: `background: var(--success-subtle)`, `border: 1px solid var(--success-border)`, `color: var(--success)`
- Completed/logged state: `background: var(--bg-elevated)`, `color: var(--text-tertiary)` — dimmed
- Hover: slightly increase background opacity

### Status Indicators (habits)
- Done: circle filled `var(--success)`, white checkmark
- At risk: circle filled `var(--warning)`, white `!`
- Neutral: circle `var(--bg-elevated)` with subtle border, dim icon
- Completed habits: entire row `opacity: 0.6`

### Nav Sidebar
- Active item: `background: #1e293b` (slightly lighter), `color: #fff`, blue right border (`2px solid var(--accent-primary)`)
- Hover: `background: #252525`, `color: #fff`
- Icons: emoji, 1.1rem

---

## 6. Dark/Light Mode — Implementation Notes

**For future implementation.** Not in scope for initial ship.

### Approach
1. Define all theme colors as CSS custom properties on `:root`
2. Override them on `[data-theme="light"]` selector
3. Toggle by setting `document.documentElement.dataset.theme`
4. Persist preference in `localStorage`
5. Respect `prefers-color-scheme` media query as default
6. Add a toggle button in the sidebar (sun/moon icon)

### What changes between modes
- All `--bg-*`, `--border-*`, `--text-*`, `--shadow-*` values
- Badge colors (backgrounds lighter in light mode)
- Score bar tracks and fills
- Status indicator fills

### What stays the same
- Font family, sizes, weights
- Spacing, radius, layout structure
- Component patterns (cards, badges, buttons)
- Accent hue family (violet in dark → rosewood in light)

---

## 7. Archived Theme: Clay Alt

The Clay Alt theme (rosewood on warm blush pink) is preserved as the light-mode direction. Key characteristics:

- **Font**: Nunito (Google Fonts) — rounded, friendly, 700–800 weight for titles
- **Background**: Warm blush `#fdecef` with white `#fff5f6` cards
- **Accent**: Rosewood `#9d6381` primary, deep berry `#612940` for success
- **Radius**: Larger than Bento Alt (14px cards vs 10px) — softer, more organic
- **Score bars**: 8px tall (thicker), terracotta → sage gradient in Clay, solid rosewood in Clay Alt
- **Badges**: Pill-shaped `border-radius: 10px`, pastel backgrounds
- **Shadows**: Warmer: `rgba(157, 99, 129, 0.08)`

Full CSS is archived in `web/src/routes/design/+page.svelte` under `.theme-clay-alt`.

---

## 8. Migration Checklist

When applying the Bento Alt theme to the production views:

- [ ] Create `web/src/lib/theme.css` with CSS custom properties from §2
- [ ] Update `+layout.svelte` to import `theme.css` and apply vars to `:global(body)`
- [ ] Replace all hardcoded colors in `+page.svelte` (inbox) with `var(--*)` references
- [ ] Replace all hardcoded colors in `today/+page.svelte` with `var(--*)` references
- [ ] Replace all hardcoded colors in `habits/+page.svelte` with `var(--*)` references
- [ ] Update `lib/badges.css` to use `var(--*)` references
- [ ] Update sidebar nav colors to match theme
- [ ] Verify all views render correctly with new theme
- [ ] Remove or keep the `/design` route (useful for future theme exploration)
