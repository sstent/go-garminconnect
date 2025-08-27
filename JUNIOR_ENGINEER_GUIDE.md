# Junior Engineer Implementation Guide

## Task Implementation Checklist
```markdown
**Task: [Task Name]**
- [ ] Study Python equivalent functionality
- [ ] Design Go implementation approach
- [ ] Implement core functionality
- [ ] Add comprehensive unit tests
- [ ] Handle edge cases and errors
- [ ] Benchmark performance
- [ ] Document public interfaces
- [ ] Submit PR for review
```

## Workflow Guidelines
1. **Task Selection**: 
   - Choose one endpoint or module at a time
   - Start with authentication before moving to endpoints
   
2. **Implementation Process**:
   ```mermaid
   graph LR
     A[Study Python Code] --> B[Design Go Interface]
     B --> C[Implement Core]
     C --> D[Write Tests]
     D --> E[Benchmark]
     E --> F[Document]
   ```

3. **Code Standards**:
   - Follow Go idioms and effective Go guidelines
   - Use interfaces for external dependencies
   - Keep functions small and focused
   - Add comments for complex logic
   - Write self-documenting code

4. **Testing Requirements**:
   - Maintain 80%+ test coverage
   - Test all error conditions
   - Verify boundary cases
   - Include concurrency tests

5. **PR Submission**:
   - Include implementation details in PR description
   - Add references to Python source
   - Document test cases
   - Note any potential limitations

## Common Endpoint Implementation Template
```go
func (c *Client) GetEndpoint(ctx context.Context, params url.Values) (*ResponseType, error) {
    // 1. Build request URL
    // 2. Handle authentication
    // 3. Send request
    // 4. Parse response
    // 5. Handle errors
    // 6. Return structured data
}
```

## Troubleshooting Tips
- When stuck, compare with Python implementation
- Use debug logging for API calls
- Verify token refresh functionality
- Check response status codes
- Consult FIT specification documents
