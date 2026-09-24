---
name: sloper-design-review
description: Evaluate the Sloper console against its Apple-inspired Liquid Glass design guide with static checks, Chrome MCP screenshots, and a concise scorecard.
---

# Sloper design review

Use this skill when changing or reviewing the Sloper dashboard's visual system.

## Read first

1. Read `dashboard/design-system/SLOPER-GLASS.md`.
2. Read `dashboard/design-system/design-contract.json`.
3. Inspect the current `dashboard/frontend/app/globals.css`, `components/app-shell.tsx`, and the main route.

## Static gate

From `dashboard/frontend` run:

```bash
npm run evaluate:design
```

Treat a failed contract as a blocker. A warning about a linear gradient is acceptable only when it is the loading skeleton; gradients must not style the canvas, cards, or navigation.

## Chrome MCP visual gate

Use a fresh isolated browser context and inspect the running dashboard at both:

- `1440 × 900` desktop
- `390 × 844` mobile

Check the following states:

1. initial instance/setup surface;
2. connected overview;
3. Events activity surface;
4. Issues empty/filter state;
5. Roadmap live-versus-planned state;
6. light and dark themes;
7. mobile navigation drawer open and closed.

Use semantic snapshots before screenshots. Wait for real API state or a visible empty state; do not inject events or mock transport data.

## Apple comparison

Use Chrome MCP to compare the current UI with:

- [Apple homepage](https://www.apple.com/)
- [Apple product page](https://www.apple.com/in/iphone-18-pro/)
- [Apple Store](https://www.apple.com/in/store)
- [OpenDesign glassmorphism guide](https://github.com/nexu-io/open-design/tree/main/design-systems/glassmorphism)
- [OpenDesign Apple guide](https://github.com/nexu-io/open-design/tree/main/design-systems/apple)

The comparison is about restraint and material behavior, not copying Apple product content:

- The canvas should be solid Apple gray/white/black rather than an ambient blue-violet wash.
- Liquid Glass should be concentrated in the shell, menus, sticky chrome, temporary overlays, and floating selectors/status capsules; each should show a translucent fill, blur/saturation depth, a light-catching edge, and a soft floating shadow.
- The sidebar, top bar, and popover may use distinct opacity tokens so they read as a physical stack, but no component should stack a second translucent utility background over the recipe.
- Data panels should be solid and quiet.
- Blue should be scarce and reserved for actions, links, focus, and selection.
- Borders and surface steps should do most of the depth work; shadows should remain soft and occasional.
- The expanded brand lockup should contain only the raven mark and lowercase `sloper`; the collapsed rail should show the raven alone.
- No single screenshot should be accepted if it looks like a generic neon glass dashboard.

## Scorecard

Report:

| Area | Weight | Pass condition |
| --- | ---: | --- |
| Foundations and tokens | 25 | Required semantic tokens and no ambient gradient tokens |
| Glass placement | 25 | Liquid Glass is visible in chrome/overlays, not every content card, and its edge/depth recipe is legible |
| Hierarchy and rhythm | 20 | Solid surfaces, clear type hierarchy, restrained blue and shadows |
| Accessibility | 20 | Focus, contrast, keyboard drawer behavior, reduced motion/transparency |
| Responsive behavior | 10 | Mobile drawer and dense data remain usable at 390px |

A visual review is not a pixel-diff against Apple. It is a reasoned comparison against the documented principles. Include the exact commit, routes inspected, viewport sizes, static-check output, and any remaining blockers.
