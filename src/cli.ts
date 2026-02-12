#!/usr/bin/env node

// src/cli.ts
//
// CLI entry point for the Google Docs MCP Server.
//
// Usage:
//   google-docs-mcp          Start the MCP server (default)
//   google-docs-mcp auth     Run the interactive OAuth flow

import { logger } from './logger.js';

const command = process.argv[2];

if (command === 'auth') {
  // Run the interactive OAuth flow and exit
  const { runAuthFlow } = await import('./auth.js');
  try {
    await runAuthFlow();
    logger.info('Authorization complete. You can now start the MCP server.');
    process.exit(0);
  } catch (error: any) {
    logger.error('Authorization failed:', error.message || error);
    process.exit(1);
  }
} else {
  // Default: start the MCP server
  const { startServer } = await import('./server.js');
  await startServer();
}
