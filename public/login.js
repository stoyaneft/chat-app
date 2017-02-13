Vue.component('chat-item', {
        props: ['chatUid'],
        template: '<li>{{ chatUid }}</li>',
});

Vue.component('chat-item', {
  props: ['id'],
  template: '<li>{{ id }}</li>'
})

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
        chatUids: [9, 10],
        chats: new Map(/*chatId: members*/),
        addedUser: '',
        activeChat: null,
    },

    created: function() {
         $(".button-collapse").sideNav();
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            console.log('msg ', msg)
            if (msg.type == 'error') {
                Materialize.toast(msg.message, 2000);
            }
            if (msg.type == 'loginSuccessful') {
                self.onLoginSuccessful();
            }
            if (msg.type == 'chatCreationSuccessful') {
                self.onChatCreationSuccessful(msg.chatUid);
            }
            if (this.joined && msg.type == 'message') {
                self.chatContent += '<div class="chip">'
                        + '<img src="' + self.gravatarURL(msg.email) + '">' // Avatar
                        + msg.username
                    + '</div>'
                    + emojione.toImage(msg.message) + '<br/>'; // Parse emojis

                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            }
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        email: this.email,
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
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
            // this.joined = true;

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
                type: 'createChat'
            }));
        },
        onChatCreationSuccessful: function(chatUid) {
            this.chats.set(chatUid, []);
            this.chatUids.push(chatUid);
            this.activeChat = chatUid;
            console.log(this.chatUids)
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});
