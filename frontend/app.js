console.log("app.js LOADED");
function restaurantApp() {
  return {
    restaurants: [],
    tables: [],
    selectedRestaurantId: null,
    selectedTableId: null,
    order: null,

    async init() {
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

      this.selectedRestaurantId = id;
      this.tables = []; // Clear old tables while loading

      try {
        const res = await fetch(`/api/v1/restaurants/${id}/tables`);
        const resp = await res.json();
        this.tables = resp.data;
      } catch (err) {
        console.error('Failed to fetch tables:', err);
        this.tables = [];
      }
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
      }
    },

    async fetchOrder(id) {
      if (id == null) return

      try {
        const res = await fetch(`/api/v1/orders/${id}`)
      } catch(err) {
        console.error('Failed to fetch order data: ', err)
      }
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
