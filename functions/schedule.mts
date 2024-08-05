export const handler = async () => {
    try {
      console.log("start req of https://almanack.data.spotlightpa.org/api-background/cron");
      let res = await fetch("https://almanack.data.spotlightpa.org/api-background/cron");
      if (res.ok) {
        console.log("res ok");
        return {
          statusCode: 200,
          body: JSON.stringify({ message: "Done" }),
        };
      }
      console.error("bad res", res.status, res.statusText);
    } catch (e) {
      console.error("could not connect", e);
    }
    return {
      statusCode: 502,
      body: JSON.stringify({ message: "Could not connect" }),
    };
  };
