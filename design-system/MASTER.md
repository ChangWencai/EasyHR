# EasyHR Design System

## Overview
Modern enterprise HR SaaS dashboard with glassmorphism aesthetic.

## Pattern
- **Name:** Enterprise Gateway
- **Style:** Glassmorphism + Clean Minimalism
- **Mood:** Professional, trustworthy, modern

## Colors
| Role | Hex | Usage |
|------|-----|-------|
| Primary | `#7C3AED` | Buttons, links, active states |
| Primary Light | `#A78BFA` | Hover states, accents |
| Primary Dark | `#5B21B6` | Pressed states |
| Background | `#FAFBFC` | Page background |
| Surface | `#FFFFFF` | Cards, panels |
| Glass | `rgba(255,255,255,0.8)` | Glassmorphism cards |
| Text Primary | `#1F2937` | Headings, primary text |
| Text Secondary | `#6B7280` | Descriptions, labels |
| Text Muted | `#9CA3AF` | Placeholders, hints |
| Border | `#E5E7EB` | Card borders, dividers |
| Success | `#10B981` | Success states |
| Warning | `#F59E0B` | Warning states |
| Error | `#EF4444` | Error states |

## Typography
- **Font Family:** Inter (Google Fonts)
- **Heading XL:** 28px / 700 / 1.2
- **Heading L:** 22px / 700 / 1.3
- **Heading M:** 18px / 600 / 1.4
- **Body:** 14px / 400 / 1.5
- **Caption:** 12px / 400 / 1.4

## Spacing
- **XS:** 4px
- **SM:** 8px
- **MD:** 16px
- **LG:** 24px
- **XL:** 32px
- **2XL:** 48px

## Border Radius
- **SM:** 8px
- **MD:** 12px
- **LG:** 16px
- **XL:** 24px
- **Full:** 9999px (pills)

## Shadows
- **SM:** `0 1px 2px rgba(0,0,0,0.05)`
- **MD:** `0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -1px rgba(0,0,0,0.06)`
- **LG:** `0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05)`
- **XL:** `0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04)`
- **Glow Primary:** `0 0 20px rgba(124,58,237,0.3)`

## Effects
### Glassmorphism
```css
background: rgba(255, 255, 255, 0.8);
backdrop-filter: blur(12px);
-webkit-backdrop-filter: blur(12px);
border: 1px solid rgba(255, 255, 255, 0.2);
```

### Hover Effects
```css
transition: all 0.2s ease-out;
transform: translateY(-2px);
box-shadow: 0 8px 16px -4px rgba(0,0,0,0.1);
```

## Animations
- **Micro-interactions:** 150-200ms
- **Page transitions:** 300-400ms
- **Easing:** `cubic-bezier(0.4, 0, 0.2, 1)` (ease-out)

## Icons
- **Library:** Element Plus Icons
- **Size SM:** 16px
- **Size MD:** 20px
- **Size LG:** 24px
- **Size XL:** 32px

## Anti-patterns
- No emojis as icons
- No excessive animations
- No dark mode by default
- No hardcoded colors outside palette
