# Sloper glass design guide

## Purpose

Sloper's console is an operator surface, not a marketing page. It borrows Apple's restraint, material cues, and information rhythm without copying Apple's product chrome or pretending to be an Apple product page.

The design goal is **quiet glass**: translucent chrome where it helps navigation and focus, solid neutral surfaces for operational data, and imagery or state—not decorative gradients—carrying visual interest.

## Apple reference observations

The live Apple pages were inspected with Chrome MCP. The useful patterns were consistent across the homepage, product pages, and Store:

- Global navigation is a thin, quiet bar: solid or lightly translucent neutral, small type, generous horizontal space, and almost no visible elevation.
- The page canvas alternates between `#ffffff`, Apple gray (`#f5f5f7`), and deliberate black chapters. It does not wash every surface in blue or violet.
- Product imagery and material rendering provide the color and drama. UI chrome stays quiet.
- Glass is concentrated in navigation, menus, floating purchase/control capsules, and temporary overlays. Content cards are generally solid, tonal surfaces.
- Blue is reserved for links, selection, and decisive actions. Secondary accents are not scattered across the interface.
- Controls use purposeful geometry: 8–12px fields and buttons, 16–18px cards, and occasional capsule controls for a single primary action.
- Depth is restrained: borders and surface steps do most of the work; shadows are soft and occasional.
- Focus and selection use a single blue signal. Hover states brighten or shift subtly; they do not glow.
- On product pages, a local navigation/purchase capsule can float over the chapter, while the global navigation recedes. A dense operator app can keep its sidebar, but should apply the same quiet layering.

Reference surfaces inspected:

- [Apple homepage](https://www.apple.com/)
- [iPhone 18 Pro product page](https://www.apple.com/in/iphone-18-pro/)
- [Apple Store](https://www.apple.com/in/store)
- [OpenDesign glassmorphism package](https://github.com/nexu-io/open-design/tree/main/design-systems/glassmorphism)
- [OpenDesign Apple package](https://github.com/nexu-io/open-design/tree/main/design-systems/apple)

## Design principles

### 1. Solid data, glass chrome

Use the glass material for the shell, menus, sticky headers, and focused overlays. Use solid `panel`, `panel-2`, and `panel-3` surfaces for metrics, tables, activity, and code. A dashboard should not look like a stack of translucent cards.

### 2. One accent, one purpose

`--color-accent` is Apple action blue. Use it for links, active navigation, focus, and the primary action. Status colors communicate state; they are not decorative second accents.

### 3. Surface rhythm before effects

Alternate solid canvas and panel tones to create hierarchy. A soft border and a small shadow are enough. Do not add ambient radial gradients, violet washes, or glow fields to ordinary content.

### 4. Chrome disappears

The sidebar and top bar should be legible and calm. Their translucency is for depth and continuity while scrolling, not for making every control float.

### 5. Data is the foreground

Issue titles, run status, event messages, and health state carry the visual weight. Decorative effects must never reduce contrast or compete with status.

## Token contract

The canonical tokens live in [`dashboard/frontend/app/globals.css`](../frontend/app/globals.css). The names are semantic; components should consume variables rather than repeat raw colors.

| Role | Light | Dark | Use |
| --- | --- | --- | --- |
| `--color-base` | `#f5f5f7` | `#000000` | App canvas |
| `--color-panel` | `#ffffff` | `#1d1d1f` | Solid cards and data surfaces |
| `--color-panel-2` | `#f5f5f7` | `#262629` | Inputs, quiet controls, secondary bands |
| `--color-panel-3` | `#ffffff` | `#2a2a2c` | Hover/raised control state |
| `--color-edge` | `#d2d2d7` | `#424245` | Hairline containment |
| `--color-edge-2` | `#86868b` | `#6e6e73` | Stronger boundary/hover |
| `--color-ink` | `#1d1d1f` | `#f5f5f7` | Primary text |
| `--color-ink-dim` | `#424245` | `#d2d2d7` | Supporting text |
| `--color-ink-faint` | `#6e6e73` | `#a1a1a6` | Metadata and captions |
| `--color-accent` | `#0071e3` | `#2997ff` | Links, focus, active state |
| `--color-glass` | translucent white | translucent graphite | Shell and overlay chrome |

The complete machine-readable contract is [`design-contract.json`](design-contract.json).

## Component recipes

### Shell and navigation

- Use a 64px desktop header, a 20px collapsed rail, and a single accent indicator for the active route.
- Use `glass-chrome` for the sidebar, header, popovers, and mobile drawer.
- Keep navigation labels visible on mobile; never persist a desktop rail state into the mobile drawer.
- Keep the drawer keyboard-contained and restore focus to its trigger on close.

### Panels

- Use `panel` for content. Default radius: 18px.
- Use `panel-hover` only for cards that navigate or expose a meaningful action.
- Use a solid surface, a quiet border, and at most one soft shadow. Do not put a gradient or glow behind every panel.
- Use `glass-hero` only when a single feature intentionally acts as a chapter or focal surface; it remains tonal, not a colored wash.

### Controls

- Primary actions use `--color-accent`, white text, and a capsule/pill radius when the action is singular and decisive.
- Secondary actions use `--color-panel-2` and `--color-edge`.
- Inputs are solid, border-led fields with a 4px accent focus ring.
- Icon-only controls require an accessible name and a visible focus state.

### Data and status

- Status badges use semantic colors, a small dot, and short literal labels.
- Charts need a textual summary and keyboard-reachable data table; the visual chart is not the only representation.
- Empty states explain what will appear and offer the next real action. Planned capabilities must say **Planned** and **Not implemented yet**.

### Motion

- Use 150–180ms for hover/focus and 220–280ms for entering surfaces.
- Use a strong ease-out curve: `cubic-bezier(0.28, 0, 0.22, 1)`.
- Do not animate from `scale(0)`; if scale is used, start at `0.9` or higher.
- Honor `prefers-reduced-motion` and `prefers-reduced-transparency`.

## Do / do not

### Do

- Start with a solid neutral canvas.
- Use glass for shell and overlays where it has a functional purpose.
- Keep blue scarce and meaningful.
- Use real borders, surface steps, and compact spacing before shadows.
- Let status, data, and product/runtime information carry the visual hierarchy.
- Test light and dark themes at desktop and mobile widths.

### Do not

- Do not use ambient radial gradients as a general dashboard background.
- Do not make every card translucent, glowing, or elevated.
- Do not use violet/blue gradients to imply depth in ordinary content.
- Do not copy Apple product language, imagery, or product claims into Sloper.
- Do not claim a planned capability is live.
- Do not remove focus rings, reduced-motion handling, or accessible chart equivalents for visual polish.

## Evaluation

Run the static contract check from the frontend package:

```bash
npm run evaluate:design
```

Then use Chrome MCP for a visual pass at 1440×900 and 390×844. The companion [`sloper-design-review`](../../.agents/skills/sloper-design-review/SKILL.md) skill describes the evidence and scorecard. A passing static check is necessary but not sufficient: the visual pass must confirm that the canvas is quiet, glass is limited to chrome, and operational data remains the foreground.
