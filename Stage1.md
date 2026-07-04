V1

This involves me making it usable. It should contain some core features:
1. Detect & Triage Issues
    1. Provide Relevant analysis of an Issue (be it bug or a feature)
    2. Files meant to change
    3. Analysis / Plan of implementaion + any required changes (should be able to accommodate my comments too)(would be semi-auto, requires permission before implemetation)

2. Implement the Plan and Create a PR
    1. Create and manage branch, Implement and solve the issue + be able to accept new changes made via comment.
    2. Also self analyze the PR (via a diff instance) and fix the suggested changes

3. Log the shell outputs

-------------------------------------------------------------------------------------------------------------------------------------------------------------------

Technical Requirements:
1. Design a good unified logger. 
2. Use the agents(pi) sdk or whatever to interact with it, to get the relevant logs and such
3. Cache the issues and pr's, comments and all.
4. Do the whole scheduler thing.




Doesnt Do:
1. Does not log the agents(pi) outputs + tool calls (will have to look into agents sdk or something)
2. Does not support multiple repo at once (I know a bummer. I would like to first make it good, than make it convinient)