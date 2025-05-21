FROM python:3.10-slim

# Install ffmpeg and other dependencies
RUN apt-get update && \
    apt-get install -y ffmpeg && \
    apt-get clean

# Set working directory
WORKDIR /src

# Copy and install requirements
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy your bot code
COPY . .

# Run the bot
CMD ["python", "bot.py"]
