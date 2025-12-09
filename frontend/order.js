
function orderApp() {
    return {
        message: "hi",
        orderId: null,
        order: null,
        isSuccess: false,

        CURRENCIES_WITH_NO_CENTS: ['sek'],

        async init () {
            console.log("order.js is LOADED")
            orderId = this.getOrderIdFromParam()
            this.getOrderDetails(orderId)

            this.checkSuccessParam()
            console.log(this.isSuccess)
        },

        async getOrderDetails(orderId) {
            if (orderId == null) return

            try {
                const res = await fetch(`/api/v1/orders/${this.orderId}`)
                const resJson = await this.raiseForStatus(res)
                this.order = resJson.data
            } catch(err) {
                console.error("failed to fetch order details: ", err)
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

            const data = await res.json()
            return data;
        },

        getOrderIdFromParam() {
            const params = new URLSearchParams(window.location.search);
            this.orderId = params.get("orderId")
            return this.orderId
        },

        checkSuccessParam() {
            const params = new URLSearchParams(window.location.search);
            this.isSuccess = true ? (params.get("success") === "true") : false
        },

        centsToFloat(amountInCents) {
            if (this.CURRENCIES_WITH_NO_CENTS.includes(this.order.currency)) return amountInCents
            if (amountInCents == null) return
            return (amountInCents / 100).toFixed(2)
        },
    }
}
