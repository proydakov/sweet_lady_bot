## TRY ME

Telegram bot with sweet girls).
Test me... https://web.telegram.org/#/im?p=@sweet_lady_bot

## BACKEND

    sudo adduser backend
    sudo mkdir /opt/sweet_lady_bot
    sudo chown -R backend:backend /opt/sweet_lady_bot

## SYSTEMD

    sudo cp /home/ubuntu/repository/sweet_lady_bot/systemd/sweet_lady_bot.service /etc/systemd/system/sweet_lady_bo.service
    sudo systemctl start sweet_lady_bot.service
    sudo systemctl enable sweet_lady_bot.service
