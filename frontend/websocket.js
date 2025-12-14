console.log("websocket.js LOADED");
function websocketApp() {
  return {
    restaurants: [],
    tables: [],
    selectedRestaurantId: null,
    selectedTableId: null,
    order: null,
    menu: null,
    tipAmount: 0.00,

    socket: null,

    SUCCESS_URL: null,
    CANCEL_URL: null,
    loadingPayment: false,
    CURRENCIES_WITH_NO_CENTS: ['sek'],

    async init() {
      this.SUCCESS_URL = `${this.getBaseURL()}/frontend/order.html?success=true&orderId=`;
      this.CANCEL_URL = `${this.getBaseURL()}/frontend/index.html?cancel=true`;

      this.checkSuccessParam();
      this.itemAddedToast = new bootstrap.Toast(
        document.getElementById('itemAddedToast'),
        {
          autohide: true,
          delay: 750
        }
      );
      try {
        const res = await fetch('/api/v1/restaurants');
        const response = await res.json();
        this.restaurants = response.data.restaurants;
      } catch (err) {
        console.error('Failed to fetch restaurants:', err);
      }
    },

    async selectRestaurant(id) {
      if (this.selectedRestaurantId === id) return

      this.order = null
      this.menu = null
      this.selectedTableId = null

      this.selectedRestaurantId = id;
      this.tables = []; // Clear old tables while loading

      try {
        const res = await fetch(`/api/v1/restaurants/${id}/tables`);
        const resp = await res.json();
        this.tables = resp.data;
      } catch (err) {
        console.error('Failed to fetch tables:', err);
        this.tables = [];
        return
      }

      this.fetchMenu(this.selectedRestaurantId)
    },

    async selectTable(id) {
      if (this.selectedTableId === id) return 

      this.selectedTableId = id

      try {
        const res = await fetch(`/api/v1/orders/current?tableId=${id}`)
        const resJson = await res.json()
        orderId = resJson.data.id
        console.log("latest order for table is: ", orderId)
      } catch (err) {
        console.error('Failed to fetch current order for table: ', err)
        return
      }

      this.fetchOrder(orderId)
      this.joinOrderWebsocket(orderId)
    },

    joinOrderWebsocket(orderId) {
      if (orderId === null) return

      if (this.socket != null) {
        this.leaveOrderWebsocket()
      }

      const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
      const host = window.location.host

      this.socket = new WebSocket(`${protocol}://${host}/api/v1/orders/${orderId}/ws`)

      this.socket.addEventListener("open", () => {
          console.log("connected to order: ", orderId)
      })

      this.socket.addEventListener("message", (event) => this.handleReceivedMsg(event));
    },

    handleReceivedMsg(event) {
      try {
          const parsedMsg =JSON.parse(event.data)

          if (parsedMsg.type === "error") throw new Error(parsedMsg.data)

          this.order = parsedMsg.data
          this.tipAmount = this.centsToFloat(parsedMsg.data.tip_amount_in_cents)
      } catch (e) {
          console.error("couldn't parse message data: ", event.data, " error: ", e)
      }
    },

    sendMessage(type, data) {
        const msg = {
            "type": type, 
            "data": data
        }
        
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(msg));
        }
    },

    leaveOrderWebsocket() {
      this.socket.close(1000, "user left ws")
      this.socket = null
    },

    async fetchMenu(id) {
      if (id == null) return

      try {
         const res = await fetch(`/api/v1/restaurants/${this.selectedRestaurantId}/menu/items`)
         const resJson = await res.json()
         this.menu = resJson.data
      } catch(err) {
        console.error("Failed to fetch menu: ", err)
      }
    },

    async fetchOrder(id) {
      if (id == null) return

      try {
        const res = await fetch(`/api/v1/orders/${id}`)
        const resJson = await res.json()

        this.order = resJson.data
        this.setTipAmount(this.order.tip_amount_in_cents)
      } catch(err) {
        console.error('Failed to fetch order data: ', err)
      }
    },

    async initPayment() {
      try {
        this.loadingPayment = true

        const res = await fetch(`/api/v1/orders/${this.order.id}/payments`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "success_url": this.SUCCESS_URL + this.order.id,
            "cancel_url": this.CANCEL_URL
          })
        })
        await this.raiseForStatus(res)

        const resJson = await res.json()
        checkoutUrl = resJson.data.url
        window.location.href = checkoutUrl
      } catch(err) {
        console.error("failed to create checkout session: ", err)
        this.loadingPayment = false
      }
    },

    async raiseForStatus(res) {
      if (!res.ok) {
        let message;
        try {
          const data = await res.json();
          message = data?.message || JSON.stringify(data);
        } catch {
          message = await res.text();
        }
        throw new Error(`HTTP ${res.status}: ${message}`);
      }
      return res;
    },

    checkSuccessParam() {
      const params = new URLSearchParams(window.location.search);

      if (params.get("success") === "true") {
        this.showSuccessToast();
      }

      if (params.get("cancel") === "true") {
        this.showCancelToast();
      }
    },

    showSuccessToast() {
      const toastEl = document.getElementById('paymentSuccessToast');
      const toast = new bootstrap.Toast(toastEl);
      toast.show();
    },

    showCancelToast() {
      const toastEl = document.getElementById('paymentCancelToast');
      const toast = new bootstrap.Toast(toastEl);
      toast.show();
    },

    showItemAddedToast() {
      this.itemAddedToast.show();
    },

    setTipAmount(amountInCents) {
      this.tipAmount = this.centsToFloat(amountInCents)
    },

    centsToFloat(amountInCents) {
      if (this.CURRENCIES_WITH_NO_CENTS.includes(this.order.currency)) return amountInCents

      if (amountInCents == null) return
      return (amountInCents / 100).toFixed(2)
    },

    floatToCents(amountInFloat) {
      if (this.CURRENCIES_WITH_NO_CENTS.includes(this.order.currency)) return amountInFloat
      if (amountInFloat == null) return
      return Math.round(amountInFloat * 100)
    },

    getBaseURL() {
      return window.location.origin;
    },

    countryCodeToFlagEmoji(code) {
      return code
        .slice(0, 2)
        .toUpperCase()
        .replace(/./g, char =>
          String.fromCodePoint(char.charCodeAt(0) + 127397)
        );
    }
  };
}
