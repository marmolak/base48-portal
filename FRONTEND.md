# Frontend Specification — Base48 Member Portal

Source of truth for all frontend decisions. When adding or modifying UI, follow these rules.

## Architecture

- **CSS framework**: Tailwind CSS v4 (CDN, loaded in `layout.html`)
- **Shared component CSS**: `web/static/css/admin.css` — buttons, badges, modals, text utilities
- **Page-specific CSS**: Inline `<style>` blocks in templates — only for styles unique to that page
- **Templates**: Go `html/template`, located in `web/templates/`

### Rule: No CSS duplication

If a style is used on more than one page, it belongs in `admin.css`. Page-specific `<style>` blocks must only contain styles unique to that template.

## Color Palette

### Buttons (solid backgrounds)

| Token           | Hex       | Usage                                    |
|-----------------|-----------|------------------------------------------|
| blue-primary    | `#2196F3` | Primary actions (`.btn-primary`)         |
| gray-primary    | `#6b7280` | Secondary/view actions (`.btn-secondary`, `.btn-view`) |
| red-primary     | `#f44336` | Destructive actions (`.btn-danger`)      |
| gray-dark       | `#4b5563` | Hover states for gray buttons            |

### Badges (pastel backgrounds, dark text)

| Variant          | Background | Text      | Usage                               |
|------------------|------------|-----------|---------------------------------------|
| green            | `#dcfce7`  | `#166534` | Success, accepted, active_member      |
| red              | `#fee2e2`  | `#991b1b` | Danger, suspended, memberportal_admin |
| yellow           | `#fef3c7`  | `#92400e` | Warning, awaiting                     |
| blue             | `#dbeafe`  | `#1e40af` | Info, roles (default), council_member |
| orange           | `#ffedd5`  | `#9a3412` | in_debt role                          |
| gray             | `#f3f4f6`  | `#6b7280` | Rejected, exmember, disabled          |
| purple           | `#ede9fe`  | `#5b21b6` | Special/sync badges                   |
| indigo           | `#e0e7ff`  | `#3730a3` | Fallback for unknown roles            |

### Text utilities

| Token           | Hex       | Usage                                    |
|-----------------|-----------|------------------------------------------|
| text-negative   | `#f44336` | Negative balance (bold)                  |
| text-positive   | `#4CAF50` | Positive balance                         |
| text-muted      | `#999`    | Placeholder / empty text                 |

## Components

### Badges (`.badge`)

Pill-shaped, pastel-background badges. Used for status indicators, roles, tags.
All badges use the same base class; never use inline Tailwind for badge styling.

```css
.badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 10px;
    border-radius: 9999px;   /* pill shape */
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;      /* NEVER wraps */
}
```

**Semantic variants** (status indicators):
- `.badge-success` — green pastel (enabled, active, OK)
- `.badge-danger` — red pastel (disabled, error, suspended)
- `.badge-warning` — yellow pastel (not linked, awaiting)

**State variants** (member states — used with `badge-{{ .State }}`):
- `.badge-accepted` — green
- `.badge-awaiting` — yellow
- `.badge-suspended` — red
- `.badge-rejected` — gray
- `.badge-exmember` — gray

**Role variants** (used with `badge-role badge-role-{{ .RoleName }}`):
- `.badge-role` — blue (fallback for unknown roles)
- `.badge-role-active_member` — green
- `.badge-role-memberportal_admin` — red
- `.badge-role-council_member` — blue
- `.badge-role-in_debt` — orange

**Generic color variants** (for misc use):
- `.badge-gray`, `.badge-yellow`, `.badge-red`, `.badge-blue`, `.badge-green`, `.badge-orange`, `.badge-purple`, `.badge-indigo`

**Usage patterns:**

State badge (auto-colored by state name):
```html
<span class="badge badge-{{ .DBUser.State }}">{{ .DBUser.State }}</span>
```

Role badges (auto-colored by role name, with fallback):
```html
<div class="badge-group">
    {{ range .Roles }}
    <span class="badge badge-role badge-role-{{ . }}">{{ . }}</span>
    {{ end }}
</div>
```

**Badge groups** — always wrap multiple badges in `.badge-group`:
```css
.badge-group {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
}
```

### Buttons (`.btn`)

```
.btn {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 500;
    text-decoration: none;
    display: inline-block;
}
```

**Sizes:**
- Default: `padding: 6px 12px; font-size: 12px;`
- `.btn-sm`: `padding: 4px 8px; font-size: 11px;`

**Variants:**
- `.btn-primary` — blue (`#2196F3`), primary actions
- `.btn-secondary` — gray (`#6b7280`), cancel/clear actions
- `.btn-danger` — red (`#f44336`), destructive actions
- `.btn-view` — gray (`#6b7280`), "view" and other neutral row-level actions

**Rule: Action buttons in tables must be visually consistent.** All row-level action buttons (Zobrazit, Manage Roles, etc.) use `.btn .btn-sm .btn-view`.

### Modals (`.modal`)

Standard pattern for popup dialogs:

```html
<div id="myModal" class="modal" style="display:none;">
    <div class="modal-content">
        <span class="close" onclick="closeMyModal()">&times;</span>
        <h2>Title</h2>
        <!-- content -->
    </div>
</div>
```

### Tables

Admin list pages use custom `.some-table` classes with consistent structure:
- Green header (`#4CAF50`) for main admin tables (admin_users)
- Gray header (`#f9fafb`) for detail/sub-tables
- `border-collapse: collapse`, `1px solid #ddd` borders

### Text Utilities

- `.text-negative` — red, bold (negative balances)
- `.text-positive` — green (positive balances)
- `.text-muted` — gray `#999` (empty/placeholder text)
- `.text-link` — styled link within tables

## Guidelines

1. **Badges never wrap.** Always use `white-space: nowrap`.
2. **Multiple badges need a `.badge-group` wrapper** for consistent spacing.
3. **Table action buttons must be consistent.** Use the same variant for all row-level actions.
4. **Shared styles live in `admin.css`.** Don't redefine `.btn`, `.badge`, `.modal` in templates.
5. **Page-specific styles stay in templates.** Filter forms, custom tables, project-specific layouts.
6. **Tailwind is available** for layout and one-off styling. Prefer it for pages that don't need custom components (logs, settings, profile).
