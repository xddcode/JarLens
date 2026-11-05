# ADR 0004: Adopt the OpenIDEA Islands design direction

- Status: Accepted
- Date: 2026-11-05

## Context

JarLens needs a familiar, efficient interface for long code-reading sessions. The
initial generic dark welcome page did not define enough of the final workbench.
The product owner selected an IntelliJ IDEA-style tool-window and editor workflow,
with OpenIDEA as the primary visual reference and support for multiple themes.

## Decision

Use OpenIDEA's Islands visual language for the workbench: a 40-pixel titlebar,
compact activity stripe, 240-pixel project window, editor island, breadcrumbs,
and status bar. Ship working Islands Dark and Islands Light themes from the same
semantic Viva token set. The default accent is blue rather than the initial teal.

The current titlebar control switches themes immediately. Persistence, OS theme
synchronization, additional palettes, and separate editor schemes are later work
with their own tests.

## Alternatives

A custom consumer-style visual direction would differentiate the welcome page but
would not improve the core code-reading workflow. Copying JetBrains proprietary
resources is not permitted. Reusing open-source assets is allowed only with their
license and attribution.

## Consequences

Design and review cover both built-in themes, populated and empty states, keyboard
focus, and window bounds. A single custom titlebar integrates application actions
with native window behavior; a separate operating-system titlebar is not shown.
The interface remains recognizably JarLens and carries no JetBrains affiliation.
Future themes must implement the complete token contract rather than changing only
the background.
