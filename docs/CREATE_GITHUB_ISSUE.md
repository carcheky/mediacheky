# How to Create the Radarr/Sonarr Configuration GitHub Issue

## Automated Method (Recommended)

Use the GitHub CLI to create the issue automatically:

```bash
gh issue create \
  --title "Implement comprehensive Radarr and Sonarr configuration interface in MediaCheky" \
  --body-file docs/ISSUE_RADARR_SONARR_CONFIG.md \
  --label "enhancement,radarr,sonarr,ui,configuration,documentation" \
  --assignee "github-copilot"
```

## Manual Method

If you prefer to create the issue manually through the GitHub web interface:

1. **Navigate to Issues**
   - Go to: https://github.com/carcheky/mediacheky/issues
   - Click "New Issue"

2. **Set Issue Title**
   ```
   Implement comprehensive Radarr and Sonarr configuration interface in MediaCheky
   ```

3. **Copy Issue Body**
   - Open `docs/ISSUE_RADARR_SONARR_CONFIG.md`
   - Copy the entire contents
   - Paste into the issue description field

4. **Add Labels**
   - `enhancement`
   - `radarr`
   - `sonarr`
   - `ui`
   - `configuration`
   - `documentation`

5. **Assign**
   - Assignee: `@github-copilot` (or relevant developer)

6. **Create Issue**
   - Click "Submit new issue"

## Quick Copy-Paste Command

For quick access, here's a one-liner to display the issue content:

```bash
cat docs/ISSUE_RADARR_SONARR_CONFIG.md
```

## Issue Reference

The issue includes:

- ✅ **Comprehensive documentation reference** (`docs/RADARR_SONARR_FEATURES.md`)
- ✅ **4-phase implementation plan** with estimated timelines
- ✅ **Detailed UI/UX recommendations** with ASCII mockups
- ✅ **Technical specifications** for backend and frontend
- ✅ **Success criteria** for each phase
- ✅ **Testing requirements** (unit, integration, E2E)
- ✅ **Documentation requirements** (user and developer)
- ✅ **Security and performance considerations**
- ✅ **Open questions** for discussion

## After Creating the Issue

1. **Link to Pull Request**
   - Reference this PR in the issue
   - Reference the issue in this PR

2. **Add to Project Board** (if applicable)
   - Add to MediaCheky development roadmap
   - Assign to appropriate milestone

3. **Notify Team**
   - Share issue link with development team
   - Discuss priority and timeline

## Related Files

- **Feature Documentation**: `docs/RADARR_SONARR_FEATURES.md`
- **Issue Template**: `docs/ISSUE_RADARR_SONARR_CONFIG.md`
- **This Guide**: `docs/CREATE_GITHUB_ISSUE.md`

---

**Note**: The issue has been prepared with all necessary details. Review the content in `docs/ISSUE_RADARR_SONARR_CONFIG.md` before creating the issue to ensure it meets your requirements.
