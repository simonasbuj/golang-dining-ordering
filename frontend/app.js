console.log("app.js LOADED");
function restaurantApp() {
  return {
    restaurants: [],
    tables: [],
    selectedRestaurantId: null,
    selectedTableId: null,
    order: null,
    menu: null,
    tipAmount: 0.00,

    SUCCESS_URL: null,
    CANCEL_URL: null,
    loadingPayment: false,

    async init() {
      this.SUCCESS_URL = `${this.getBaseURL()}/frontend/index.html?success=true`;
      this.CANCEL_URL = `${this.getBaseURL()}/frontend/index.html?cancel=true`;

      this.checkSuccessParam();
      this.itemAddedToast = new bootstrap.Toast(document.getElementById('itemAddedToast'));
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

    async addItemToOrder(itemId) {
      if (itemId == null) return

      try {
        const res = await fetch(`/api/v1/orders/${this.order.id}/items`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            item_id: itemId
          })
        })
        await this.raiseForStatus(res)

        const respJson = await res.json()

        this.order = respJson.data
        this.setTipAmount(this.order.tip_amount_in_cents)
        this.showItemAddedToast()
      } catch(err) {
        console.error('Failed to add item to an order: ', err)
      }
    },

    async removeItemFromOrder(itemId) {
      if (itemId == null) return

      try {
        const res = await fetch(`/api/v1/orders/${this.order.id}/items`, {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            item_id: itemId
          })
        })
        await this.raiseForStatus(res)

        const respJson = await res.json()
        this.order = respJson.data
        this.setTipAmount(this.order.tip_amount_in_cents)
      } catch(err) {
        console.error('Failed to remove item from an order: ', err)
      }
    },

    async editTip(tipAmount) {
      tipAmountInCents = this.floatToCents(tipAmount)
      if (tipAmountInCents == null || tipAmountInCents == this.order.tip_amount_in_cents || tipAmountInCents < 0) return

      try {
        const res = await fetch(`/api/v1/orders/${this.order.id}`, {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "tip_amount_in_cents": tipAmountInCents
          })
        })
        await this.raiseForStatus(res)

        const resJson = await res.json()
        this.updateCurrentOrder(resJson.data)
      } catch(err) {
        console.error("failed to update tip amount: ", err)
      }
    },

    async lockOrder() {
      if (this.order.id == null) return
      
      try {
        const res = await fetch(`/api/v1/orders/${this.order.id}`, {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "status": "locked"
          })
        })
        await this.raiseForStatus(res)

        const resJson = await res.json()
        this.updateCurrentOrder(resJson.data)
      } catch(err) {
        console.error("failed to lock order: ", err)
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
            "success_url": this.SUCCESS_URL,
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
      toast.show();
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

    updateCurrentOrder(updatedOrder) {
      if (updatedOrder == null) return
      this.order = updatedOrder
      this.setTipAmount(this.order.tip_amount_in_cents)
    },

    setTipAmount(amountInCents) {
      this.tipAmount = this.centsToFloat(amountInCents)
    },

    centsToFloat(amountInCents) {
      if (amountInCents == null) return
      return (amountInCents / 100).toFixed(2)
    },

    floatToCents(amountInFloat) {
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
