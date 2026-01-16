# Task 105: Public UX — Alpine.js Lightbox and Carousel

**Milestone:** Admin UI  
**Points:** 2 (6 hours)  
**Dependencies:** 100  
**Branch:** `feat/lightbox`  
**Labels:** `public`, `alpine`, `ux`

## Description
Add a lightbox modal with Alpine.js for full-screen photo viewing and keyboard-navigable carousel within albums.

## Acceptance Criteria
- [ ] Click photo opens lightbox overlay
- [ ] Lightbox shows full-resolution image
- [ ] Previous/Next buttons navigate within album
- [ ] Keyboard controls: Arrow keys (prev/next), Escape (close)
- [ ] Focus trap while lightbox open
- [ ] Lightbox closes on background click
- [ ] Accessible (ARIA labels, focus restoration)

## Files to Add/Modify
- `web/templates/public/lightbox.html` — lightbox component
- `web/static/lightbox.js` — Alpine.js component (if separate)
- `web/templates/public/album_gallery.html` — integrate lightbox

## Alpine.js Component
```html
<div x-data="lightbox()" @keydown.escape.window="close()" @keydown.arrow-left.window="prev()" @keydown.arrow-right.window="next()">
    <!-- Photo Grid -->
    <div id="photo-grid">
        <img @click="open({{ .Index }})" src="{{ .URL }}" class="cursor-pointer">
    </div>
    
    <!-- Lightbox Modal -->
    <div x-show="isOpen" x-cloak class="fixed inset-0 bg-black bg-opacity-90 z-50 flex items-center justify-center">
        <button @click="close()" class="absolute top-4 right-4 text-white">✕</button>
        <button @click="prev()" class="absolute left-4 text-white">‹</button>
        <button @click="next()" class="absolute right-4 text-white">›</button>
        <img :src="currentPhoto.url" class="max-w-full max-h-full">
    </div>
</div>

<script>
function lightbox() {
    return {
        isOpen: false,
        currentIndex: 0,
        photos: {{ .Photos | json }},
        get currentPhoto() {
            return this.photos[this.currentIndex];
        },
        open(index) {
            this.currentIndex = index;
            this.isOpen = true;
            document.body.style.overflow = 'hidden';
        },
        close() {
            this.isOpen = false;
            document.body.style.overflow = '';
        },
        prev() {
            this.currentIndex = (this.currentIndex - 1 + this.photos.length) % this.photos.length;
        },
        next() {
            this.currentIndex = (this.currentIndex + 1) % this.photos.length;
        }
    }
}
</script>
```

## Tests Required
- [ ] Manual test: click photo opens lightbox
- [ ] Manual test: arrow keys navigate photos
- [ ] Manual test: escape key closes lightbox
- [ ] Manual test: background click closes lightbox
- [ ] Manual test: focus returns to clicked photo on close

## PR Checklist
- [ ] Lightbox works on mobile (touch swipe optional for MVP)
- [ ] Keyboard navigation functional
- [ ] Focus trap prevents tabbing outside modal
- [ ] ARIA roles and labels added (role="dialog", aria-label)
- [ ] Lightbox closes on Escape and background click
- [ ] Manual test: no console errors

## Git Workflow
```bash
git checkout -b feat/lightbox
# Implement Alpine.js lightbox
# Manual test in browser
git add web/templates/public/ web/static/
git commit -m "feat: add Alpine.js lightbox and carousel for photos"
git push origin feat/lightbox
# Open PR: "Implement lightbox modal with keyboard navigation"
```

## Notes
- Use Alpine.js x-data, x-show, @click, @keydown for reactivity
- Preload adjacent images for smooth navigation (optional)
- For MVP, swipe gestures can be deferred
- Ensure lightbox works with HTMX-loaded photos (dynamic binding)
