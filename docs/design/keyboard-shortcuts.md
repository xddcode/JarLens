# Keyboard and command model

This document defines the intended command model as features land. Only commands
visible in the running application are considered implemented.

## Conventions

Use Command on macOS and Control on Windows/Linux for primary shortcuts. Every
shortcut has a named command in the palette and an accessible menu or control.
Shortcuts do not override text editing expectations.

| Command | Planned shortcut | Availability |
| --- | --- | --- |
| Command palette | Primary+Shift+P | M1 |
| Open archive | Primary+O | M1 |
| Close tab | Primary+W | M1 |
| Find in source | Primary+F | M1 |
| Search workspace | Primary+Shift+F | M2 |
| Go to line | Primary+G | M1 |
| Next/previous match | F3 / Shift+F3 | M1 |
| Cancel current task | Escape, when task focused | M1 |
| Source/engine diff | No default initially | M3 |

Tab traversal, arrows, Enter, Space, Home/End, and Page keys follow Viva's native
component behavior. Escape dismisses the topmost transient surface; it does not
silently discard a workspace.

A PR introducing a command updates this table, accessible labels, tests, and user
documentation in the same change.
