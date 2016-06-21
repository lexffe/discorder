## Builtin Commands

| Name | Description | Category | Args |
| --- | --------- | ------- | ---- |
| server_notifications_settings | Change notifications settings for a server | Discord | <ul><li>server:string - Server to change settings on</li></ul> | 
| commands | Opens up the command window with all commands available | Hidden | <ul></ul> | 
| move_cursor | Moves cursor in specified direction | hidden | <ul><li>direction:string</li><li>amount:int</li><li>word:boolean</li></ul> | 
| erase | Erase text | hidden | <ul><li>direction:string</li><li>amount:int</li><li>words:boolean</li></ul> | 
| servers | Opens up the server window | Windows | <ul></ul> | 
| help | Opens up the help window | Windows | <ul></ul> | 
| scroll | Scrolls currently active view | hidden | <ul><li>direction:string</li><li>amount:int</li></ul> | 
| select | Select the currently highlighted element | hidden | <ul></ul> | 
| toggle | Toggles the currently highlited element | hidden | <ul></ul> | 
| autocomplete_selection | Changes the autocomplete selection | hidden | <ul><li>amount:boolean - The amoount to change in</li></ul> | 
| clear_log | Clear the logbuffer | Utils | <ul></ul> | 
| reload_theme | Reloads the current theme | Utils | <ul></ul> | 
| theme_window | Select a theme | Windows | <ul></ul> | 
| delete_message | Deletes a message | Discord | <ul><li>last_yours:boolean - If true deletes last message you sent</li><li>last_any:boolean - If true deletes last message anyone sent</li><li>message:string - Specify a message id</li><li>channel:string - Specify a channel id</li></ul> | 
| status | Updates your discord status | Discord | <ul><li>game:string - What game you should appear playing as</li><li>idle:int - How long you've been idle in seconds</li></ul> | 
| send_message | Sends a message, (This is a stub, not implemented :() | Discord | <ul></ul> | 
| initiate_conversation | Initiate a conversation | Discord | <ul><li>user:string - User to intiate a conversation with</li></ul> | 
| set_nick | Sets your nickname on a server (if possible) | Discord | <ul><li>name:string - The nickname you will set (empty to reset)</li><li>server:string - Server to set the nickname on</li><li>user:string - Specify a user, leave empty for youself</li></ul> | 
| pin_message | Pins a message, (This is a stub, not implemented :() | Discord | <ul><li>message:string - The message that will be pinned</li><li>channel:string - The message that will be pinned</li></ul> | 
| back | Closes the active window | hidden | <ul></ul> | 
| close_windows | Closes all windows | hidden | <ul></ul> | 
| discorder_settings | Change settings |  | <ul><li>short_guilds:boolean - Displays a mini version of guilds in message view</li><li>hide_nicknames:boolean - Shows usernames instead of nicknames if true</li><li>time_format_full:string - Sets the full time format</li><li>time_format_short:string - Sets the short time format (for messages on the same day)</li></ul> | 
| change_tab | Change tab | hidden | <ul><li>tab:int</li><li>change:int</li></ul> | 
| remove_tab | Removes the active tab | Utils | <ul></ul> | 
| rename_tab | Renames the currently selected tab | Utils | <ul><li>name:string - The name you want to give</li></ul> | 
| gen_command_table | Generates command docs | Utils | <ul></ul> | 
| quit | Quit discorder |  | <ul></ul> | 