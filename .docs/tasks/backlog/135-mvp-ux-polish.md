# Task 135: MVP Polish — Responsive Design and Accessibility

**Milestone:** Metrics & Polish  
**Points:** 2 (7 hours)  
**Dependencies:** 105  
**Branch:** `feat/ux-polish`  
**Labels:** `ux`, `accessibility`, `polish`

## Description
Final UX polish for MVP: ensure responsive design works on all devices, improve accessibility (ARIA, keyboard nav), and add loading states.

## Acceptance Criteria
- [ ] Mobile-first responsive design (320px to 1920px+)
- [ ] Touch-friendly buttons and controls (min 44px tap targets)
- [ ] Loading indicators for all async operations
- [ ] Accessible forms (labels, error messages, focus states)
- [ ] Keyboard navigation works throughout
- [ ] ARIA labels for interactive elements
- [ ] Color contrast meets WCAG AA standards
- [ ] Empty states are informative and actionable

## Files to Add/Modify
- `web/static/styles.css` — responsive tweaks, loading spinners
- `web/templates/components/loading.html` — loading indicator component
- `web/templates/layouts/base.html` — accessibility meta tags
- All templates — add ARIA labels and semantic HTML

## Responsive Breakpoints
- **Mobile**: 320px - 767px (1 column grid)
- **Tablet**: 768px - 1023px (2 column grid)
- **Desktop**: 1024px+ (3-4 column grid)

## Accessibility Checklist
- [ ] All images have alt text
- [ ] Forms have associated labels
- [ ] Focus indicators visible
- [ ] ARIA roles on interactive elements (buttons, modals, etc.)
- [ ] Skip links for keyboard users
- [ ] No reliance on color alone for information

## Loading States
- HTMX requests show spinner or progress bar
- Upload progress visible (already in task 055)
- Button disabled during submission

## Tests Required
- [ ] Manual test: navigate entire site with keyboard only
- [ ] Manual test: use screen reader (NVDA, VoiceOver)
- [ ] Manual test: resize browser from 320px to 1920px
- [ ] Manual test: tap targets on mobile (real device)
- [ ] Automated: run Lighthouse accessibility audit (score > 90)

## PR Checklist
- [ ] Responsive on iPhone SE, iPad, desktop
- [ ] Lighthouse accessibility score > 90
- [ ] No console errors or warnings
- [ ] Loading states consistent across app
- [ ] Focus styles visible and attractive

## Git Workflow
```bash
git checkout -b feat/ux-polish
# Polish responsive design and accessibility
# Test on multiple devices
# Run Lighthouse audit
git add web/static/ web/templates/
git commit -m "feat: polish responsive design and accessibility for MVP"
git push origin feat/ux-polish
# Open PR: "Polish UX for MVP: responsive design and accessibility"
```

## Notes
- Use Tailwind responsive classes (`sm:`, `md:`, `lg:`)
- Test with real devices when possible (not just browser resize)
- Screen reader testing is critical for accessibility
- Loading indicators prevent user confusion during slow operations
