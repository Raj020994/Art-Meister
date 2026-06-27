# Backend API Route & Field Usage Report

## Architecture
- **All Go backend calls** go through `service/api.js` → `fetch(NEXT_PUBLIC_BACKEND_URL + endpoint)`
- **Response envelope**: `{ Success: boolean, Data: <payload>, Message?: string }`
- **Internal Next.js routes** (`/app/api/`) handle uploads, emails, password reset tokens

---

## AUTH

### POST /auth/login → `loginUser(formData)`
**Sends:** `{ email, password }`

**Used by:** `sign-in/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.id` | ✅ | redirect to `/u/{id}` |
| Everything else | ❌ | |

---

### POST /auth/users → `signUpUser(formData)`
**Sends:** `{ name, email, password, confirmPassword }`

**Used by:** `sign-up/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | redirect to `/u/{ID}` |
| Everything else | ❌ | |

> **Note:** login returns lowercase `id`, signup returns uppercase `ID` — inconsistency

---

### GET /auth/me → `getCurrUser()`
**Sends:** (none — cookie auth)

**Used by:** `Navbar.jsx` (x2), `UserMenu.jsx` (via zustand store)
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | profile link, user identification |
| `Data.Name` | ✅ | UserMenu display |
| `Data.Email.String` | ✅ | UserMenu display |
| `Data.Role` | ✅ | admin menu gating in Navbar |
| `Data.Status` | ✅ | banned checks, profile display |
| `Data.Username.Valid` + `.String` | ✅ | profile username, edit checks |
| `Data.Description.Valid` + `.String` | ✅ | profile about section |
| `Data.Batch.Valid` + `.String` | ✅ | profile details |
| `Data.Image.Valid` + `.String` | ✅ | avatar, banner fallback |
| `Data.BannerImage.Valid` + `.String` | ✅ | profile banner |
| `Data.SocialLinks.instagram` | ✅ | profile display |
| `Data.SocialLinks.youtube` | ✅ | profile display |
| Everything else | ? | stored in zustand, any consumer may access |

> **Keep all fields** — the full user object is stored in zustand `useAuthStore` and multiple pages consume different parts.

---

### POST /auth/logout → `logOutUser()`
**Sends:** (none)
**Response:** `{ Success }` only

---

## USERS

### GET /auth/users → `getAllUser()`
**Sends:** (none)

**Used by:** `admin/page.jsx` → `ManageAccount.jsx`
| Response Field (per user) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | actions, keys |
| `Data[].Name` | ✅ | display |
| `Data[].Email` | ✅ | display, search |
| `Data[].Role` | ✅ | badge, role toggle |
| `Data[].Status` | ✅ | badge, approve/reject/ban |
| `Data[].Image.Valid` + `.String` | ✅ | avatar display |
| Everything else | ❌ | description, batch, social links, etc. not used |

---

### GET /auth/main-users → `getAllApprovedUser()`
**Sends:** (none)

**Used by:** `components/Artist.jsx`
| Response Field (per user) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | link |
| `Data[].Name` | ✅ | display |
| `Data[].Role` | ✅ | display |
| `Data[].Description.String` | ✅ | truncated bio |
| `Data[].Image.String` | ✅ | avatar |
| `Data[].SocialLinks.instagram` | ✅ | social icon |
| `Data[].SocialLinks.youtube` | ✅ | social icon |
| Everything else | ❌ | batch, status, email, banner, username |

---

### PATCH /auth/users/{id} → `updateUser(id, data)`
**Sends:** `{ username, description, batch, image, banner_image, social: { instagram, youtube } }`

**Used by:** `onboarding/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data` (full) | ✅ | stored in zustand via `setUser(updatedUser?.Data)` |
| `Data.ID` | ✅ | redirect |
| Everything else | ✅ | indirectly via store |

---

### GET /auth/users/{id} → `usrById(id)`
**Not used by any page** — dead code

---

## ART

### POST /art/ → `createArt(formData)`
**Sends:** `{ name, description, url, tags: string[] }`

**Used by:** `art/create/CreateArt.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | redirect |
| Everything else | ❌ | |

---

### GET /art → `getAllArt()`
**Sends:** (none)

**Used by:** `art/page.jsx`
| Response Field (per art) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | key, link |
| `Data[].UserID` | ✅ | link to artist |
| `Data[].Image` | ✅ | display |
| `Data[].Name` | ✅ | display |
| `Data[].Description.Valid` + `.String` | ✅ | tooltip |
| `Data[].Tags` | ✅ | tag badges |
| Everything else | ❌ | Status, CreatedAt, etc. |

---

### GET /art/{id} → `getArtById(id)`
**Sends:** (none)

**Used by:** `art/create/CreateArt.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.Name` | ✅ | form edit prefill |
| `Data.Description.String` | ✅ | form edit prefill |
| `Data.Tags` | ✅ | categories prefill |
| `Data.Image` | ✅ | preview |
| Everything else | ❌ | |

---

### PATCH /art/{id} → `updateArt(id, data)`
**Sends:** `{ name?, description?, tags? }` (changed fields only)

**Used by:** `art/create/CreateArt.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | redirect |
| Everything else | ❌ | |

---

### DELETE /art/{id} → `deleteArt(id)`
**Sends:** (none)

**Used by:** `u/[userId]/page.jsx`, `u/[userId]/[artid]/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | filter from local list (profile page) |
| `Success` | ✅ | check & redirect (detail page) |
| Everything else | ❌ | |

---

### GET /art/u/{userId} → `getAllArtistArt(id)`
**Sends:** (none)

**Used by:** `u/[userId]/page.jsx`
| Response Field (per art) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | key, delete action |
| `Data[].Image` | ✅ | card display |
| `Data[].Name` | ✅ | card display |
| `Data[].Description.String` | ✅ | hover tooltip |
| Everything else | ❌ | Tags, Status, UserID, CreatedAt |

---

### GET /art/u/profile/{userId} → `getArtistProfile(id)`
**Sends:** (none)

**Used by:** `u/[userId]/page.jsx`
| Response Field (User) | Used? | Where |
|---|---|---|
| `Data.User.ID` | ✅ | admin actions |
| `Data.User.Name` | ✅ | display |
| `Data.User.Role` | ✅ | role display |
| `Data.User.Status` | ✅ | approve/ban badge |
| `Data.User.Username.Valid` + `.String` | ✅ | header |
| `Data.User.Image.Valid` + `.String` | ✅ | avatar, banner fallback |
| `Data.User.BannerImage.Valid` + `.String` | ✅ | header banner |
| `Data.User.Batch.Valid` + `.String` | ✅ | detail |
| `Data.User.Description.String` | ✅ | about section |
| `Data.User.SocialLinks.instagram` | ✅ | social icon |
| `Data.User.SocialLinks.youtube` | ✅ | social icon |

| Response Field (Art) | Used? | Where |
|---|---|---|
| `Data.Art[].ID` | ✅ | key |
| `Data.Art[].Image` | ✅ | card |
| `Data.Art[].Name` | ✅ | display |
| `Data.Art[].Description.String` | ✅ | tooltip |
| Everything else (on art) | ❌ | Tags, Status, UserID, CreatedAt |

---

### GET /art/p/{usrId}/{id} → `getArtProfileById({ usrId, id })`
**Sends:** (none)

**Used by:** `u/[userId]/[artid]/page.jsx`

**Also stored in zustand** `useArtStore` → wrapped as `{ data: ..., cachedAt: ... }`

| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | (via store cache) |
| `Data.Status` | ✅ | gating view for rejected |
| `Data.UserID` | ✅ | back link |
| `Data.Username.String` | ✅ | artist sidebar |
| `Data.UserImage.String` | ✅ | artist sidebar avatar |
| `Data.Image` | ✅ | main display |
| `Data.Name` | ✅ | title |
| `Data.Description.String` | ✅ | about section |
| Everything else | ❌ | Tags, CreatedAt, etc. |

---

### GET /art/pending-art → `getPendingArt()`
**Sends:** (none)

**Used by:** `admin/page.jsx` → `ApproveArt.jsx`
| Response Field (per art) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | approve/reject action |
| `Data[].Name` | ✅ | display |
| `Data[].Image` | ✅ | image preview |
| `Data[].Status` | ✅ | badge |
| `Data[].Description.Valid` + `.String` | ✅ | description display |
| `Data[].Tags` | ✅ | tag badges |
| `Data[].CreatedAt` | ✅ | submission date |
| Everything else | ❌ | UserID, etc. |

---

### PATCH /admin/arts/{id}/status?status={status} → `changeArtStatus(id, status)`
**Sends:** (query param only)

**Used by:** `ApproveArt.jsx`, `u/[userId]/[artid]/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | filter from list (ApproveArt) |
| `Data.Status` | ✅ | toast message |
| Everything else | ❌ | |

---

## EVENTS

### GET /event/ → `getAllEvents()`
**Sends:** (none)

**Used by:** `event/page.jsx`, `components/Events.jsx`
| Response Field (per event) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | key, link |
| `Data[].Name` | ✅ | title |
| `Data[].Description.String` | ✅ | description |
| `Data[].Image.String` | ✅ | card/featured image |
| `Data[].EventDate` | ✅ | date display |
| `Data[].Venue.String` | ✅ | venue display |
| `Data[].Status` | ✅ | online/offline badge (Events.jsx) |
| Everything else | ❌ | BannerImage, etc. |

---

### GET /event/{id} → `getEventById(id)`
**Sends:** (none)

**Used by:** `event/[eventId]/page.jsx`, `admin/events/create/CreateEventPage.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | (various) |
| `Data.Name` | ✅ | title |
| `Data.Description.String` | ✅ | about section |
| `Data.Image.String` | ✅ | logo |
| `Data.BannerImage.String` | ✅ | banner |
| `Data.EventDate` | ✅ | date, state calculation |
| `Data.Venue.String` | ✅ | venue |
| `Data.Status` | ✅ | (admin edit form) |
| Everything else | ❌ | CreatedAt, etc. |

---

### POST /event/ → `createEvent(eventData)`
**Sends:** `FormData` — `{ name, description, venue, status, date, LogoUrl, bannerUrl }`

**Used by:** `admin/events/create/CreateEventPage.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | redirect |
| Everything else | ❌ | |

---

### PATCH /event/{id} → `updateEvent(id, eventData)`
**Sends:** `{ name?, description?, date?, status?, venue?, image?, bannerImage? }`

**Used by:** `admin/events/create/CreateEventPage.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.ID` | ✅ | redirect |
| Everything else | ❌ | |

---

### DELETE /event/{id} → `deleteEvent(id)`
**Sends:** (none)

**Used by:** `event/[eventId]/page.jsx`, `EventCard.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Success` | ✅ | check |
| `Message` | ✅ | toast (detail page) |
| Everything else | ❌ | |

---

### POST /event/{id}/join → `joinEvent(id)`
**Sends:** (none)

**Used by:** `event/[eventId]/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Success` | ✅ | redirect to my-events |
| Everything else | ❌ | |

---

### GET /event/u/{id} → `getMyEvent(id)`
**Sends:** (none)

**Used by:** `event/[eventId]/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Success` | ✅ | sets `isRegistered = true` |
| Everything else | ❌ | |

---

### GET /event/all → `getMyAllEvents()`
**Sends:** (none)

**Used by:** `my-events/page.jsx` → `EventCard.jsx`
| Response Field (per event) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | key, link |
| `Data[].Name` | ✅ | title |
| `Data[].EventDate` | ✅ | date badge |
| `Data[].Image.String` | ✅ | card image |
| `Data[].Description.String` | ✅ | description |
| `Data[].Venue.String` | ✅ | venue label |
| Everything else | ❌ | Status, BannerImage |

---

### GET /event/{id}/attendees → `getEventAttendees(id)`
**Sends:** (none)

**Used by:** `event/[eventId]/member/page.jsx`
| Response Field (per attendee) | Used? | Where |
|---|---|---|
| `Data[].ID` | ✅ | key, remove action |
| `Data[].Name` | ✅ | display |
| `Data[].Username.String` | ✅ | @handle display |
| `Data[].Email` | ✅ | display |
| `Data[].Image.String` | ✅ | avatar |
| `Data[].Batch.String` | ✅ | batch display |
| `Data[].Status` | ✅ | badge |
| Everything else | ❌ | Role, SocialLinks, etc. |

---

### DELETE /event/{id}/attendee/{userid} → `deleteEventAttendee({ id, userid })`
**Sends:** (none)

**Used by:** `event/[eventId]/member/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data` | ✅ | filtered out of local list |
| Everything else | ❌ | |

---

## ADMIN

### PATCH /admin/users/{id}/status?status={...} or ?role={...} → `changeUserRoleStatus(id, { status?, role? })`
**Sends:** (query param only)

**Used by:** `ManageAccount.jsx`, `u/[userId]/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Data.Status` | ✅ | update local user state (profile page) |
| `Success` | ✅ | check (ManageAccount) |
| Everything else | ❌ | |

---

## INTERNAL NEXT.JS ROUTES (BFF)

### POST /api/upload
**Sends:** `FormData` — `{ file: File }`

**Used by:** `CreateArt.jsx`, `CreateEventPage.jsx`, `onboarding/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `success` | ✅ | check |
| `url` | ✅ | send to create/update endpoint |
| Everything else | ❌ | |

---

### POST /api/send-email
**Sends:** `{ to, resetURL }`

**Used by:** `forgot-password/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `success` | ✅ | toast (via `sent?.Success`) |
| Everything else | ❌ | |

---

### POST /api/auth/forgot-password
**Sends:** `{ email }`

**Used by:** `forgot-password/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Success` | ✅ | check |
| `Data.email` | ✅ | pass to send-email |
| `Data.resetLink` | ✅ | pass to send-email |
| Everything else | ❌ | |

---

### POST /api/auth/reset-password
**Sends:** `{ token, password }`

**Used by:** `reset-password/page.jsx`
| Response Field | Used? | Where |
|---|---|---|
| `Success` | ✅ | redirect |
| Everything else | ❌ | |

---

## HIGH-VALUE CUTS (Biggest payload reductions)

These endpoints return many fields but only a fraction are used:

| Endpoint | Returned Fields | Fields Used | Waste |
|---|---|---|---|
| `GET /art` | ~10+ fields/art | 6 | `Status`, `CreatedAt`, `UpdatedAt`, etc. per artwork — significant at scale |
| `GET /art/u/{id}` | ~10+ fields/art | 4 | Tags, Status, UserID, timestamps unused |
| `GET /auth/users` | ~10+ fields/user | 6 | Description, Batch, SocialLinks, Username, BannerImage etc. all unused |
| `GET /auth/main-users` | ~10+ fields/user | 7 | Batch, Status, Email, BannerImage, Username unused |
| `GET /event/` | ~8+ fields/event | 6 | BannerImage, CreatedAt, etc. unused |
| `GET /event/all` | ~8+ fields/event | 6 | Same — BannerImage unused |
| `GET /art/pending-art` | ~10+ fields/art | 7 | UserID, etc. unused |
| `POST /art/` (response) | Full art object | 1 | Only `Data.ID` used |
| `POST /event/` (response) | Full event object | 1 | Only `Data.ID` used |
| `POST /auth/login` (response) | Full user object | 1 | Only `Data.id` used |
| `POST /auth/users` (response) | Full user object | 1 | Only `Data.ID` used |
| `PATCH /art/{id}` (response) | Full art object | 1 | Only `Data.ID` used |
| `PATCH /event/{id}` (response) | Full event object | 1 | Only `Data.ID` used |
