document.addEventListener("alpine:init", async () => {
    Alpine.store('top', {
        state: {},
        loaded: false,
        async fetchData() {
            const response = await fetch("/api/state");
            const data = await response.json();
            this.state = data;
            this.loaded = true;
        },
        init() {
            // this.fetchData();
        }
    })

    async function runPeriodically() {
        await Alpine.store('top').fetchData(); // Wait for fetchData to complete
        setTimeout(runPeriodically, 2000); // Then wait an additional 1000 ms before next run
    }

    await runPeriodically();

    // setInterval(() => {
    //     Alpine.store('top').fetchData();
    // }, 2000);
})

console.log("index.js loaded")