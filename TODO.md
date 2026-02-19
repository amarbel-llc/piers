 Here's what's needed:                                                                                                
                                                                                                                       
  You need to do manually (I can't do these for you):                                                                  
                                                                                                                       
  1. Go to https://console.cloud.google.com/                                                                           
  2. Create a project (or use an existing one)                                                                         
  3. Enable these 3 APIs: Google Docs API, Google Sheets API, Google Drive API
  4. Configure OAuth consent screen as "External", add your email as a test user                                       
  5. Create an OAuth client ID (Desktop app type)                                                                      
  6. Copy the Client ID and Client Secret                                                                              

  Once you have those, I can handle the rest:
  - Run the auth flow (npx google-docs-mcp auth)
  - Add the MCP server config to your Claude Code settings

  Do you already have a Google Cloud project with these APIs enabled, or do you need to set that up from scratch?
