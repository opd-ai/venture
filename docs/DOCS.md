TASK: Aggressively clean up and consolidate old reports and documentation in a repository through direct deletion.

CONTEXT: Repository contains accumulated reports and documentation that need immediate cleanup. Focus is on speed and storage recovery while preserving LLM prompt templates and de-bloating remaining documentation.

REQUIREMENTS:
- Define clear deletion criteria upfront
- Preserve LLM prompt templates and structured prompt files
- Identify and eliminate duplicate or superseded documents
- Consolidate overlapping materials
- Execute deletions directly without backup overhead
- De-bloat remaining documentation by removing unnecessary details
- Extract and preserve valuable content from deleted documents into consolidated versions

EXECUTION PLAN:

**Phase 1: Define Deletion Criteria**
1. Establish retention thresholds:
   - Age cutoff (e.g., delete anything older than X years)
   - File type priorities (drafts, duplicates, superseded versions)
   - Size targets for storage recovery
   - Active project exemptions
   
2. Define LLM prompt preservation rules:
   - PRESERVE: Markdown files with structured prompt templates (TASK, CONTEXT, INSTRUCTIONS sections)
   - PRESERVE: Files containing AI/LLM directives or system prompts
   - PRESERVE: Prompt engineering documentation and examples
   - PRESERVE: Files in directories named 'prompts/', 'agents/', or '.github/copilot-instructions/'
   - PRESERVE: Files ending with patterns like '*-prompt.md', '*-agent.md', 'copilot-instructions.md'
   - PRESERVE: All files mentioned in `.sh` shell scripts such as `loop.sh` and `loop2.sh`
   - EVALUATE: General documentation files for de-bloating (Phase 6)

**Phase 2: Quick Assessment**
3. Scan repository and identify:
   - Files exceeding age threshold
   - Draft/versioned files with finals available
   - Duplicate files (same name, similar content)
   - Obsolete project documentation
   - Oversized low-value files
   - LLM prompts to preserve (marked with PRESERVE tag)

4. Apply deletion criteria to classify:
   - DELETE NOW (clearly obsolete)
   - CONSOLIDATE (merge similar docs)
   - KEEP (active/recent)
   - PRESERVE (LLM prompts and templates)
   - DE-BLOAT (documents to streamline in Phase 6)

**Phase 3: Consolidation**
5. For CONSOLIDATE items:
   - Group related documents
   - Keep most recent or authoritative version only
   - Delete all others immediately
   
6. Create minimal archive structure for kept items:
   ```
   /active/
   /archive/
     /YYYY/
   ```

**Phase 4: Content Extraction and Update**
7. Before deletion, extract valuable content:
   - Identify unique information in documents marked for deletion
   - Extract key insights, data points, or examples not present in kept versions
   - Note critical references, links, or metadata
   
8. Update consolidated documents:
   - Merge extracted content into appropriate consolidated documents
   - Add missing sections or subsections from deleted files
   - Incorporate unique examples or case studies
   - Update tables of contents and cross-references
   - Preserve important historical context where relevant
   - Add attribution notes if multiple sources were merged
   
9. Quality check updated documents:
   - Verify all valuable content has been preserved
   - Ensure logical flow after content integration
   - Remove redundant information created by merging
   - Validate that consolidated document remains coherent

**Phase 5: Execution**
10. Delete in priority order:
    - Drafts where finals exist (after content extraction)
    - Duplicates (after verifying no unique content)
    - Files exceeding age threshold (after preserving key insights)
    - Obsolete project folders (after archiving critical information)
   
11. Move remaining files to clean structure
12. Delete empty directories

**Phase 6: Cleanup**
13. Update repository README with new structure
14. Generate summary metrics

**Phase 7: De-bloating**
15. Review files marked for DE-BLOAT:
    - Remove verbose explanations that can be condensed
    - Eliminate redundant examples and repetitive content
    - Strip unnecessary formatting and whitespace
    - Condense multi-paragraph sections to essential points
    - Remove outdated references and deprecated information
    
16. Apply de-bloating systematically:
    - Target files >500 lines first (highest impact)
    - Set reduction targets based on document type:
      * Reference docs/specs: 20-30% reduction (use 20% for critical specs, 30% for verbose ones)
      * User guides/tutorials: 30-40% reduction (use 30% for concise guides, 40% for wordy tutorials)
      * Verbose reports/analyses: 40-50% reduction (use 40% for data-heavy reports, 50% for opinion pieces)
    - Maintain technical accuracy and key details
    - Keep one clear example per concept (remove extras)
    - Preserve all section headers and structure
    
17. Quality check de-bloated documents:
    - Verify core information remains intact
    - Ensure readability is maintained or improved
    - Confirm no broken references or links
    - Validate technical accuracy of condensed content

OUTPUT FORMAT:

**Cleanup Summary Report:**
```markdown
# Repository Cleanup Summary
Date: [YYYY-MM-DD]

## Results
- Files deleted: [count]
- Storage recovered: [size in GB/TB]
- Files consolidated: [count] → [count]
- Files remaining: [count]
- LLM prompts preserved: [count]
- Files de-bloated: [count]
- Average size reduction from de-bloating: [X%]
- Content extraction updates: [count of documents updated with extracted content]

## Deletion Criteria Used
- Age threshold: [X years]
- File types targeted: [list]
- Size threshold: [if applicable]
- LLM prompt preservation rules applied

## Content Consolidation
- Documents updated with extracted content: [count]
- Unique insights preserved: [brief summary]
- Examples/case studies merged: [count]

## New Repository Structure
[Brief outline of cleaned structure]

## De-bloating Summary
- Files processed: [count]
- Total word count reduced: [original count] → [new count] ([X%] reduction)
- Largest reductions: [list top 3 files with percentages]
```

QUALITY CRITERIA:
✓ Significant storage space recovered
✓ Duplicate files eliminated
✓ Clear, simplified repository structure
✓ Only recent/active materials retained
✓ LLM prompts and templates preserved
✓ Valuable content from deleted files preserved in consolidated documents
✓ Remaining documentation streamlined through de-bloating
✓ Cleanup completed efficiently

EXECUTION CHECKLIST:
- [ ] Deletion criteria defined
- [ ] LLM prompt preservation rules established
- [ ] Age/type filters applied
- [ ] Duplicates identified
- [ ] Consolidation completed
- [ ] Content extraction from files marked for deletion
- [ ] Consolidated documents updated with extracted content
- [ ] Quality check on updated documents
- [ ] Direct deletions executed
- [ ] Empty folders removed
- [ ] Structure simplified
- [ ] De-bloating targets identified
- [ ] De-bloating completed with quality checks

EXAMPLE DELETION DECISIONS:

**Scenario 1: Version Stack**
- Files: Report_v1.pdf, Report_v2.pdf, Report_v3_DRAFT.pdf, Report_FINAL.pdf
- Action: Extract unique charts from v2 → Add to Report_FINAL.pdf → DELETE v1, v2, v3_DRAFT

**Scenario 2: Age-based**
- File: Q2_2018_Analysis.xlsx
- Age: 6+ years old
- Action: Extract key metrics summary → Add to annual_trends.md → DELETE original

**Scenario 3: Duplicates**
- Files: Meeting_Notes.docx, Meeting_Notes (1).docx, Meeting_Notes_copy.docx
- Action: Compare all versions → Merge unique action items → Keep Meeting_Notes.docx (updated) → DELETE duplicates

**Scenario 4: Project Cleanup**
- Folder: /archived_project_2017/
- Contains: 247 files, all >5 years old
- Action: Extract lessons learned → Add to project_retrospectives.md → DELETE entire folder

**Scenario 5: LLM Prompt Preservation**
- File: code-review-prompt.md
- Content: Contains "TASK:", "CONTEXT:", "INSTRUCTIONS:" sections with AI directives
- Naming: Matches '*-prompt.md' pattern
- Action: PRESERVE (structured LLM prompt template - identified by both content and naming pattern)

**Scenario 6: Documentation De-bloating**
- File: installation-guide.md (2,500 words)
- Contains: Verbose explanations, 8 redundant examples, outdated references
- Action: DE-BLOAT → Reduce to 1,200 words (keep 1-2 key examples, condense explanations)