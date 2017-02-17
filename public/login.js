'use strict';

new Vue({
    el: '#login',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        chatContent: '', // A running list of chat messages displayed on the screen
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        password: null,
        joined: false, // True if email and username have been filled in
        chatUids: [],
        chats: new Map(/*chatId: members*/),
        addedUser: '',
        activeChat: {
            uid: '',
            participants: [],
            messages: ''
        },
    },

    created: function() {
         // $(".button-collapse").sideNav();
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            console.log('msg ', msg);
            switch (msg.type) {
                case 'error':
                    console.log('ha')
                    Materialize.toast(msg.message, 2000);
                    break;
                case 'loginSuccessful':
                    self.onLoginSuccessful();
                    break;
                case 'chatCreationSuccessful':
                    self.onChatCreationSuccessful(msg.chatUid);
                    break;
                case 'chatSelectionSuccessful':
                    self.onChatSelectionSuccessful(msg.chatUid, msg.participants);
                    break;
                case 'userAddedSuccessful':
                    self.onUserAddedSuccessful(msg.username);
                    break;
                case 'sendMessage':
                    self.onSendMessage(msg);
                    break;
                case 'addedToChat':
                    self.onAddedToChat(msg.chatUid, msg.participants);
            }
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        type: "sendMessage",
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text(),
                        chatUid: this.activeChat.uid
                    }
                ));
                this.newMsg = ''; // Reset newMsg
            }
        },
        login: function () {
            if (!this.username) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            if (!this.password) {
                Materialize.toast('You must enter a password', 2000);
                return
            }
            this.username = $('<p>').html(this.username).text();
            this.password = $('<p>').html(this.password).text();

            this.ws.send(JSON.stringify({
                email: this.email,
                username: this.username,
                password: this.password,
                type: 'login'
            }));
        },
        onLoginSuccessful: function() {
            this.joined = true;
        },
        createChat: function() {
            this.ws.send(JSON.stringify({
                type: 'createChat',
                username: this.username
            }));
        },
        onChatCreationSuccessful: function(chatUid) {
            this.chats.set(chatUid, { uid: chatUid, participants: [], messages: '' });
            this.chatUids.push(chatUid);
            console.log(this.chatUids)
        },
        onChatSelection: function(e) {
            this.ws.send(JSON.stringify({
                type: 'chatSelection',
                username: this.username,
                chatUid: e.target.innerText
            }))
        },
        onChatSelectionSuccessful: function(chatUid, participants) {
            var otherMembers = participants.filter(p => p !== this.username);
            console.log(otherMembers)
            var chat = this.chats.get(chatUid);
            chat.participants = otherMembers;
            this.activeChat = chat;
            console.log(this.chats)
            console.log('active chat' , this.activeChat);
        },
        addUser: function() {
            console.log('adding');
            console.log(this.addedUser);
            if (this.activeChat.participants.indexOf(this.addedUser) !== -1 || this.addedUser === this.username) {
                Materialize.toast(`The user ${this.addedUser} is already in the chat`, 2000);
                return;
            }
            this.ws.send(JSON.stringify({
                type: 'addUser',
                username: this.addedUser,
                chatUid: this.activeChat.uid
            }));
        },
        onUserAddedSuccessful: function (addedUser) {
            var chat = this.chats.get(this.activeChat.uid);
            chat.participants.push(addedUser);
            console.log(this.activeChat.participants);
        },

        onAddedToChat: function (chatUid, participants) {
            this.chats.set(chatUid, { uid: chatUid, participants, messages: '' });
            this.chatUids.push(chatUid);
        },

        onSendMessage: function (msg) {
            console.log('received message ', msg);
            var chatMessage = '<div class="chip">'
                + '<img src="' + this.gravatarURL(msg.email) + '">' // Avatar
                + msg.username
                + '</div>'
                + emojione.toImage(msg.message) + '<br/>'; // Parse emojis
            var chat = this.chats.get(msg.chatUid);
            console.log(chat);
            chat.messages += chatMessage;
            // this.chats.set(msg.chatUid, chat);
            if (msg.chatUid === this.activeChat.uid) {
                console.log(this.activeChat)
                // this.activeChat.messages += chatMessage;
                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
                console.log(this.activeChat.messages);
            }
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});
