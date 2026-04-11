# Mobile Patterns

> [!important]
> These patterns apply to mobile development — React Native, Flutter, and native. Framework-specific examples are noted where patterns diverge.

---

## Architecture

- Separate presentation from business logic — no API calls in widgets or components
- Repository pattern for all data access: UI -> ViewModel/BLoC -> Repository -> DataSource
- One state management solution per app — do not mix approaches:
  - **Flutter:** BLoC/Cubit (preferred), Riverpod, or Provider
  - **React Native:** Zustand (preferred), Redux Toolkit, or Jotai
  - **Native iOS:** SwiftUI `@Observable`, TCA
  - **Native Android:** ViewModel + StateFlow, Compose state
- Unidirectional data flow: state flows down, events flow up
- Feature-based folder structure, not layer-based: `features/auth/`, `features/profile/`, not `screens/`, `models/`, `services/`

## Navigation

- Single source of truth for routes — define all routes in one place
- Deep linking from day one: every screen must be reachable via URL
- Type-safe route parameters — never pass raw strings between screens
- Handle back navigation explicitly — don't rely on default platform behavior for complex flows
- Preserve navigation state across app restarts for key flows

## State Management

- Distinguish UI state (modal open, scroll position) from domain state (user profile, cart items)
- UI state stays local to the widget/component — never in global store
- Domain state lives in a global store or scoped provider
- Derive computed state — never store what can be calculated
- Treat loading and error as explicit states, not boolean flags: `idle | loading | success(data) | error(message)`

## Offline-First

- Local database is the primary data source — network is the sync layer
- Write to local first, sync to server in the background
- Queue failed network requests for retry (exponential backoff, max 3 retries)
- Conflict resolution strategy — pick one and apply consistently:
  - **Last-write-wins:** simple, works for non-collaborative data
  - **Server-wins:** safest for shared data where server is authoritative
  - **Manual merge:** for collaborative or critical data, surface conflicts to the user
- Background sync with platform APIs: WorkManager (Android), BGTaskScheduler (iOS), background fetch (React Native)
- Show stale data with a "last synced" indicator rather than blocking on network

## Networking

- Single HTTP client instance with shared config (timeouts, interceptors, base URL)
- Request timeout: 30s for standard, 60s for uploads, 10s for health checks
- Retry with exponential backoff for 5xx and network errors — never retry 4xx
- Cancel in-flight requests when the user navigates away
- Cache responses where appropriate — respect `Cache-Control` headers
- Certificate pinning for sensitive endpoints (banking, auth)

## Performance

- **Cold start:** target <3 seconds to interactive content
- **Frame time:** <16ms per frame (60fps), <8ms for 120fps devices
- Lazy-load screens and heavy components — never load everything at app start
- Windowed/virtualized lists for any list >20 items: `FlatList` (RN), `ListView.builder` (Flutter), `LazyColumn` (Compose)
- Image optimization: resize to display dimensions, use WebP/AVIF, cache aggressively
- Minimize main thread work — move computation, JSON parsing, and image processing off-thread
- Profile before optimizing — use platform profilers (Flipper, DevTools, Instruments) to find actual bottlenecks

## Platform Conventions

- Follow Material Design (Android) and Human Interface Guidelines (iOS) unless the brand requires deviation
- Handle safe areas, notches, and dynamic islands — never let content render under system UI
- Support Dynamic Type (iOS) and font scaling (Android) — use relative sizes, never hardcode font px
- Respect platform-specific gestures: swipe-to-go-back (iOS), back button (Android)
- Haptic feedback for confirmations and destructive actions — use platform APIs, never vibration motor directly

## Permissions

- Request permissions just-in-time, not at app launch
- Explain why before requesting: "We need camera access to scan QR codes"
- Handle denial gracefully — degrade feature, never crash or show empty screen
- Handle "Don't ask again" state — direct user to system settings with a clear explanation
- Check permission status before using the feature, every time — don't cache permission grants

## Testing

- Unit test all business logic (ViewModels, BLoCs, repositories) — mock data sources
- Widget/component tests for complex UI interactions
- Integration tests for critical flows: login, core feature, purchase
- E2E on real devices for release validation — emulators miss real-world issues (network, performance, sensors)
- Snapshot tests for UI regression — but limit to stable components, not rapidly changing screens

## Security

- Never store sensitive data in SharedPreferences/UserDefaults — use Keychain (iOS) or EncryptedSharedPreferences (Android)
- Pin SSL certificates for auth and payment endpoints
- Obfuscate release builds (ProGuard/R8 for Android, bitcode for iOS)
- Disable screenshots on sensitive screens (banking, passwords)
- Clear sensitive data from memory when backgrounding
- Validate all deep link parameters — treat them as untrusted external input
