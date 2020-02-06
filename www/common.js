export async function PostJSON(url, Data) {
    const Request = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(Data),
    });

    const Response = await Request.json();
    if (!Response.success)
        console.error(Response.msg);

    return Response;
}
