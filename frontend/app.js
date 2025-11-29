console.log("app.js LOADED");
function restaurantApp() {
  return {
    restaurants: [],
    tables: [],
    selectedRestaurantId: null,

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
      if (this.selectedRestaurantId === id) {
        return;
      }

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
