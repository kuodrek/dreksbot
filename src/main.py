import discord
from discord.ext import commands
import asyncio
import yt_dlp
import os

intents = discord.Intents.default()
intents.message_content = True  # Required for receiving message content

bot = commands.Bot(command_prefix="!nigel!", intents=discord.Intents.all())

@bot.event
async def on_ready():
    print(f"Logged in as {bot.user}!")

@bot.command()
async def hello(ctx):
    await ctx.send("Hello, world!")

@bot.command()
async def p(ctx, url: str):
    # Check if the user is in a voice channel
    if ctx.author.voice is None:
        await ctx.send("You need to be in a voice channel to use this command.")
        return

    voice_channel = ctx.author.voice.channel

    # Connect to the voice channel
    vc = await voice_channel.connect()

    # Use yt_dlp to get audio source
    ydl_opts = {
        "format": "bestaudio",
        "quiet": True,
        "noplaylist": True,
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        info = ydl.extract_info(url, download=False)
        audio_url = info["url"]

    # Create FFmpeg audio player
    ffmpeg_options = {"options": "-vn"}

    source = discord.FFmpegPCMAudio(audio_url, **ffmpeg_options)
    vc.play(source)

    await ctx.send(f"Now playing: {info['title']}")


@bot.command()
async def s(ctx):
    if ctx.voice_client is None:
        await ctx.send("I'm not in a voice channel.")
        return

    await ctx.voice_client.disconnect()
    await ctx.send("I have left the voice channel.")

token = os.environ["DISCORD_TOKEN"]
bot.run(token)
