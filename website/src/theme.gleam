//// Shared Tailwind class tokens for a consistent look across the site:
//// a single `neutral` gray scale, and `pink` as the only accent color.

/// Page background, used on the root `<html>` element.
pub const page_bg = "bg-neutral-100 dark:bg-neutral-950"

/// Default body text color.
pub const text_default = "text-neutral-900 dark:text-neutral-100"

/// Muted/secondary text (captions, descriptions, footer text).
pub const text_muted = "text-neutral-500 dark:text-neutral-400"

/// Heading and code text color, slightly softer than the default.
pub const text_heading = "text-neutral-800 dark:text-neutral-200"

/// Raised surface background - code blocks, sidebar, section dividers.
pub const surface_bg = "bg-neutral-100 dark:bg-neutral-900"

/// Section separator background for alternating sections.
pub const section_bg = "bg-white dark:bg-neutral-900"

/// Standalone border on all sides.
pub const border = "border border-neutral-300 dark:border-neutral-700"

/// Horizontal divider below an element.
pub const border_b = "border-b border-b-neutral-300 dark:border-b-neutral-700"

/// Horizontal divider above an element.
pub const border_t = "border-t border-t-neutral-300 dark:border-t-neutral-700"

/// Accent text color for highlighted elements (no hover state).
pub const accent_text = "text-pink-600 dark:text-pink-300"

/// Full link styling: accent color with a hover state.
pub const accent_link = "text-pink-600 dark:text-pink-300 hover:text-pink-400 dark:hover:text-pink-200"

/// Accent button background.
pub const accent_bg = "bg-pink-500/30 hover:bg-pink-500/80"
