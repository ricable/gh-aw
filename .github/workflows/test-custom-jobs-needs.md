---
on: workflow_dispatch

permissions:
  contents: read

jobs:
  # First job: data collection
  collect_data:
    needs: activation
    runs-on: ubuntu-latest
    outputs:
      item_count: ${{ steps.count.outputs.count }}
    steps:
      - name: Count items
        id: count
        run: |
          echo "count=5" >> $GITHUB_OUTPUT
  
  # Second job: depends on data collection
  process_data:
    needs: collect_data
    runs-on: ubuntu-latest
    steps:
      - name: Process ${{ needs.collect_data.outputs.item_count }} items
        run: |
          echo "Processing items..."
  
  # Third job: runs after agent with data from collect_data
  report:
    needs: [agent, collect_data]
    if: always()
    runs-on: ubuntu-latest
    steps:
      - name: Generate report
        run: |
          echo "Report: processed ${{ needs.collect_data.outputs.item_count }} items"
---

# Test Custom Jobs with Dependencies

This workflow tests that custom jobs with `needs` dependencies work correctly.

**Instructions**: Just confirm the workflow compiled successfully and has the expected job dependencies.
