# Design Guide

---

## Component Patterns

- One component per file — named exports matching the filename
- Prefer composition over prop-drilling — use context or state management for deeply shared state
- Keep components focused — if a component does more than one thing, split it
- Separate presentational components from data-fetching containers

## Styling

- Use the project's configured CSS solution consistently (Tailwind, CSS Modules, etc.)
- Never use inline styles except for truly dynamic values (e.g., calculated positions)
- Design tokens (colors, spacing, typography) come from the theme — never hardcode

## Accessibility

- All interactive elements must be keyboard-navigable
- Images need `alt` text — decorative images get `alt=""`
- Form inputs need associated labels
- Color must not be the only way to convey information
- Test with screen reader when adding new interactive patterns

## Responsive

- Mobile-first approach — start with the smallest breakpoint
- Test at standard breakpoints: 375px, 768px, 1024px, 1440px
- No horizontal scrolling at any breakpoint
