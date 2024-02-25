FROM scratch

ENV GAPS_LOGIN_USERNAME=""                          \
    GAPS_LOGIN_PASSWORD=""                          \
    GAPS_HISTORY_GRADES_FILE="/history/grades.json" \
    GAPS_SCRAPER_API_URL=""                         \
    GAPS_SCRAPER_API_KEY=""

ENTRYPOINT ["/gaps-cli"]
COPY gaps-cli /

CMD ["--help"]
